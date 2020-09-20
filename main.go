package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"zoove/controllers"
	"zoove/middleware"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/gomodule/redigo/redis"
)

type HostSpotifyClientAuth struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

var UnAuthorized = errors.New("Error authorizing this guy.")
var NotFound = errors.New("Not Found")
var IncompleteRequest = errors.New("The request is incomplete. An import part is missing")

const (
	HostDeezer     = "deezer"
	HostSpotify    = "spotify"
	HostAppleMusic = "applemusic"
)

// func loadEnv() {
// 	envr := os.Getenv("ENV")
// 	err := godotenv.Load(".env." + envr)
// 	if err != nil {
// 		log.Println("Error reading the env file")
// 		log.Println(err)
// 		panic(err)
// 	}
// }

// func init() {
// 	loadEnv()
// }

var pool *redis.Pool

func main() {
	app := fiber.New()

	app.Static("/", "./client/build")
	jaeger := controllers.NewJaeger(pool)

	app.Use(middleware.ExtractInfoMetadata)
	app.Get("/api/v1.1/search", jaeger.JaegerHandler)
	app.Get("/api/v1", func(ctx *fiber.Ctx) {
		ctx.Status(http.StatusOK).Send("Hi")
	})

	app.Use(ExtractInfo)
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodOptions, http.MethodDelete},
	}))
	app.Get("/api/v1/search", EquivalentsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "13200"
	}
	app.Listen(port)
}

type ExtractedInfo struct {
	Host string
	URL  string
	ID   string
}

func ExtractInfo(ctx *fiber.Ctx) {
	rawURL := ctx.Query("track")
	if rawURL == "" {
		log.Println("the req is invalid")
		ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "The request is missing important part", "error": "Missing track query parameter"})
		return
	}

	song, err := url.QueryUnescape(rawURL)
	if err != nil {
		log.Println("Error escaping URL")
		ctx.Next(err)
	}
	parsedURL, err := url.Parse(song)
	if err != nil {
		log.Println("Error parsing URL")
		ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error getting parsing the URL", "error": err.Error()})
		return
	}

	platformHost := parsedURL.Host
	index := strings.Index(song, "?")
	sub := ""
	if index == -1 {
		sub = song
	} else {
		sub = song[:index]
	}
	midd := &ExtractedInfo{}
	// for deezer, a song is typically like this:A, https://www.deezer.com/en/track/545820622. but to
	// use the API to get song info, its like this:B, https://api.deezer.com/track/3135556.
	// the below code simply uses the url from A and turn it into B

	if platformHost == "www.deezer.com" {
		deezerID := sub[32:]
		midd.Host = "deezer"
		midd.URL = fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
		midd.ID = deezerID
	} else if platformHost == "open.spotify.com" {
		spotifyID := sub[31:]
		midd.Host = "spotify"
		midd.URL = sub
		midd.ID = spotifyID
	}

	ctx.Locals("extractedInfo", midd)
	ctx.Next()
}

func EquivalentsHandler(ctx *fiber.Ctx) {
	extracted := ctx.Locals("extractedInfo").(*ExtractedInfo)

	var track = SearchTrack{}
	if extracted.Host == HostDeezer {

		info, err := HostDeezerFetchTrackMetaData(extracted.URL)
		if err != nil {
			log.Println("Error fetching metadata from deezer")
			log.Println(err)
			if err == NotFound {
				ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "Not found", "error": "The track does not exist"})
				return
			}
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
			return
		}
		track = *info
	} else if extracted.Host == HostSpotify {
		txURL := fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), extracted.ID)
		out, err := HostSpotifyFetchMetaData(txURL)
		if err != nil {
			log.Println("Error getting hotspot stuff")
			log.Println(err)

			if err == NotFound {
				ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Not Found", "error": "Song does not exist."})
				return
			}
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
			return
		}
		track = *out
	}

	equivalents, err := FetchEquivalents(track)
	if err != nil {
		log.Println("Error getting the equivalents..")
		log.Println(err)
		ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK).JSON(fiber.Map{"message": "Found", "data": equivalents})
}

func HostDeezerExtractTitle(title string) string {
	// we want to check for the first occurence of "Feat"
	ind := strings.Index(title, "(feat")
	if ind == -1 {
		return title
	}
	out := title[:ind]
	return out
}

func FetchEquivalents(search SearchTrack) ([]*SearchedSong, error) {
	deezer, err := search.HostDeezerSearchForTrack()

	if err != nil {
		return nil, err
	}

	spotify, err := search.HostSpotifySearchForTrack()
	if err != nil {
		return nil, err
	}

	results := []*SearchedSong{}
	results = append(results, deezer, spotify)
	return results, nil
}

type SearchedSong struct {
	Title       string   `json:"title"`
	Duration    int      `json:"duration"`
	Artistes    []string `json:"artistes"`
	URL         string   `json:"url"`
	Preview     string   `json:"preview"`
	Cover       string   `json:"cover"`
	ReleaseDate string   `json:"release_date"`
	Explicit    bool     `json:"explicit"`
	Platform    string   `json:"platform"`
}

func (search *SearchTrack) HostDeezerSearchForTrack() (*SearchedSong, error) {
	track := &SearchedSong{}

	title := HostDeezerExtractTitle(search.Title)
	formattedURL := fmt.Sprintf("track:\"%s\" artist:\"%s\" album:\"%s\"", title, search.ArtisteName, search.AlbumName)
	escape := url.QueryEscape(formattedURL)
	searchURL := fmt.Sprintf("%s/search?q=%s", os.Getenv("DEEZER_API_BASE"), escape)
	output := &HostDeezerSearchResult{}
	err := MakeRequest(searchURL, output)
	if err != nil {
		return nil, err
	}

	if len(output.Data) > 0 {
		base := output.Data[0]
		track.Artistes = []string{base.Artist.Name}
		track.Cover = base.Album.Cover
		track.Duration = base.Duration * 1000 // to convert to miliseconds
		track.Explicit = base.ExplicitLyrics
		track.Preview = base.Preview
		track.ReleaseDate = ""
		track.Title = base.Title
		track.URL = base.Link
		track.Platform = "deezer"

		return track, nil
	}
	return nil, nil
}

func (search *SearchTrack) HostSpotifySearchForTrack() (*SearchedSong, error) {
	track := &SearchedSong{}
	search.IsAlbum = false
	payload := fmt.Sprintf("track:%s artist:%s", search.Title, search.ArtisteName)
	escaped := url.QueryEscape(payload)
	searchURL := fmt.Sprintf("%sv1/search", os.Getenv("SPOTIFY_API_BASE"))
	url := fmt.Sprintf("%s?q=%s&type=track", searchURL, escaped)

	t, err := GetSpotifyAuthToken()
	if err != nil {
		log.Println("Error getting token from spotify:")
		log.Println(err)
		return nil, err
	}
	token := t.(*HostSpotifyClientAuth)
	output := &HostSpotifySearchResult{}

	err = MakeSpotifyRequest(url, token.AccessToken, output)

	if err != nil {
		log.Println("Error making spotify request for search track here")
		return nil, err
	}

	if len(output.Tracks.Items[0].Artists) > 0 {
		artistes := []string{}
		for i := range output.Tracks.Items[0].Artists {
			artistes = append(artistes, output.Tracks.Items[0].Artists[i].Name)
		}
		base := output.Tracks.Items[0]
		track.Artistes = artistes
		track.Cover = base.Album.Images[0].URL
		track.Duration = base.DurationMs
		track.Explicit = base.Explicit
		track.Preview = base.PreviewURL
		track.ReleaseDate = base.Album.ReleaseDate
		track.Title = base.Name
		track.URL = base.ExternalUrls.Spotify
		track.Platform = "spotify"
		return track, nil
	}

	return nil, nil
}

func HostDeezerFetchTrackMetaData(url string) (*SearchTrack, error) {
	output := &HostDeezerTrack{}
	full := &SearchTrack{}

	err := MakeRequest(url, output)
	if err != nil {
		log.Println("Error getting song Deezer metadata")
		log.Println(err)
		return nil, err
	}
	if output.TrackPosition == output.DiskNumber {
		full.IsAlbum = false
		full.AlbumName = ""
	} else {
		full.IsAlbum = true
		full.AlbumName = output.Album.Title
		full.TrackNumber = output.TrackPosition
	}
	full.ArtisteName = output.Artist.Name
	full.Title = output.Title
	full.URL = output.Link
	full.PreviewURL = output.Preview
	full.Platform = "deezer"
	return full, nil
}

func HostSpotifyFetchMetaData(url string) (*SearchTrack, error) {
	t, err := GetSpotifyAuthToken()
	if err != nil {
		log.Println("Error getting token from spotify:")
		log.Println(err)
		return nil, err
	}
	output := &HostSpotifyTrack{}
	full := &SearchTrack{}

	token := t.(*HostSpotifyClientAuth)
	err = MakeSpotifyRequest(url, token.AccessToken, output)
	if err != nil {
		log.Println("Error making request to spotify")
		return nil, err
	}

	full.AlbumName = output.Album.Name
	full.ArtisteName = output.Album.Artists[0].Name
	full.TrackNumber = output.TrackNumber
	full.Title = output.Name
	full.Platform = "spotify"
	full.URL = output.ExternalUrls.Spotify
	return full, nil
}

func MakeSpotifyRequest(url, token string, output interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		log.Println("Error making request to the URL")
		log.Println(err)
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		log.Println("Error making HTTP request")
		log.Println(err)
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading response body")
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return UnAuthorized
	} else if res.StatusCode == http.StatusNotFound {
		return NotFound
	}

	err = json.Unmarshal(body, output)
	if err != nil {
		log.Println("Error parsing the body to JSON")
		return err
	}

	return nil
}

func GetSpotifyAuthToken() (interface{}, error) {

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifySecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", strings.NewReader(reqBody.Encode()))
	if err != nil {
		log.Fatalf("Error with spotify auth")
	}

	bearer := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifySecret))

	req.Header.Set("Authorization", "Basic "+bearer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(reqBody.Encode())))
	doRequest, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(doRequest.Body)
	if err != nil {
		return nil, err
	}
	defer doRequest.Body.Close()
	out := &HostSpotifyClientAuth{}

	err = json.Unmarshal(body, out)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return out, nil

}

func MakeRequest(url string, output interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Error making request to the URL")
		log.Println(err)
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)

	// log.Printf("Headers from deezer is: %s", res.Header)
	if err != nil {
		log.Println("Error making HTTP request")
		log.Println(err)
		return err
	}
	body, err := ioutil.ReadAll(res.Body)

	if strings.Contains(string(body), "{\"error\"") {
		return NotFound
	}
	if err != nil {
		log.Println("Error reading response body")
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return UnAuthorized
	}

	err = json.Unmarshal(body, output)
	if err != nil {
		log.Println("Error parsing the body to JSON")
		return err
	}
	return nil
}

type SearchTrack struct {
	IsAlbum     bool   `json:"is_album"`
	ArtisteName string `json:"artiste_name"`
	AlbumName   string `json:"album_name,omitempty"`
	URL         string `json:"url"`
	Title       string `json:"track_title"`
	TrackNumber int    `json:"track_number,omitempty"`
	Platform    string `json:"platform"`
	PreviewURL  string `json:"preview_url"`
}
type HostDeezerTrack struct {
	ID                    int      `json:"id"`
	Readable              bool     `json:"readable"`
	Title                 string   `json:"title"`
	TitleShort            string   `json:"title_short"`
	TitleVersion          string   `json:"title_version"`
	Isrc                  string   `json:"isrc"`
	Link                  string   `json:"link"`
	Share                 string   `json:"share"`
	Duration              int      `json:"duration"`
	TrackPosition         int      `json:"track_position"`
	DiskNumber            int      `json:"disk_number"`
	Rank                  int      `json:"rank"`
	ReleaseDate           string   `json:"release_date"`
	ExplicitLyrics        bool     `json:"explicit_lyrics"`
	ExplicitContentLyrics int      `json:"explicit_content_lyrics"`
	ExplicitContentCover  int      `json:"explicit_content_cover"`
	Preview               string   `json:"preview"`
	Bpm                   float64  `json:"bpm"`
	Gain                  float64  `json:"gain"`
	AvailableCountries    []string `json:"available_countries"`
	Contributors          []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Link          string `json:"link"`
		Share         string `json:"share"`
		Picture       string `json:"picture"`
		PictureSmall  string `json:"picture_small"`
		PictureMedium string `json:"picture_medium"`
		PictureBig    string `json:"picture_big"`
		PictureXl     string `json:"picture_xl"`
		Radio         bool   `json:"radio"`
		Tracklist     string `json:"tracklist"`
		Type          string `json:"type"`
		Role          string `json:"role"`
	} `json:"contributors"`
	Artist struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Link          string `json:"link"`
		Share         string `json:"share"`
		Picture       string `json:"picture"`
		PictureSmall  string `json:"picture_small"`
		PictureMedium string `json:"picture_medium"`
		PictureBig    string `json:"picture_big"`
		PictureXl     string `json:"picture_xl"`
		Radio         bool   `json:"radio"`
		Tracklist     string `json:"tracklist"`
		Type          string `json:"type"`
	} `json:"artist"`
	Album struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		Link        string `json:"link"`
		Cover       string `json:"cover"`
		CoverSmall  string `json:"cover_small"`
		CoverMedium string `json:"cover_medium"`
		CoverBig    string `json:"cover_big"`
		CoverXl     string `json:"cover_xl"`
		ReleaseDate string `json:"release_date"`
		Tracklist   string `json:"tracklist"`
		Type        string `json:"type"`
	} `json:"album"`
	Type string `json:"type"`
}

type HostSpotifyAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type HostSpotifyTrack struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		TotalTracks          int    `json:"total_tracks"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`
	DiscNumber       int      `json:"disc_number"`
	DurationMs       int      `json:"duration_ms"`
	Explicit         bool     `json:"explicit"`
	ExternalIds      struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string      `json:"href"`
	ID          string      `json:"id"`
	IsLocal     bool        `json:"is_local"`
	Name        string      `json:"name"`
	Popularity  int         `json:"popularity"`
	PreviewURL  interface{} `json:"preview_url"`
	TrackNumber int         `json:"track_number"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
}

type HostDeezerSearchResult struct {
	Data []struct {
		ID                    int    `json:"id"`
		Readable              bool   `json:"readable"`
		Title                 string `json:"title"`
		TitleShort            string `json:"title_short"`
		TitleVersion          string `json:"title_version"`
		Link                  string `json:"link"`
		Duration              int    `json:"duration"`
		Rank                  int    `json:"rank"`
		ExplicitLyrics        bool   `json:"explicit_lyrics"`
		ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
		ExplicitContentCover  int    `json:"explicit_content_cover"`
		Preview               string `json:"preview"`
		Md5Image              string `json:"md5_image"`
		Artist                struct {
			ID            int    `json:"id"`
			Name          string `json:"name"`
			Link          string `json:"link"`
			Picture       string `json:"picture"`
			PictureSmall  string `json:"picture_small"`
			PictureMedium string `json:"picture_medium"`
			PictureBig    string `json:"picture_big"`
			PictureXl     string `json:"picture_xl"`
			Tracklist     string `json:"tracklist"`
			Type          string `json:"type"`
		} `json:"artist"`
		Album struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			Cover       string `json:"cover"`
			CoverSmall  string `json:"cover_small"`
			CoverMedium string `json:"cover_medium"`
			CoverBig    string `json:"cover_big"`
			CoverXl     string `json:"cover_xl"`
			Md5Image    string `json:"md5_image"`
			Tracklist   string `json:"tracklist"`
			Type        string `json:"type"`
		} `json:"album"`
		Type string `json:"type"`
	} `json:"data"`
	Total int `json:"total"`
}

type HostSpotifySearchResult struct {
	Tracks struct {
		Href  string `json:"href"`
		Items []struct {
			Album struct {
				AlbumType string `json:"album_type"`
				Artists   []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height int    `json:"height"`
					URL    string `json:"url"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"release_date_precision"`
				TotalTracks          int    `json:"total_tracks"`
				Type                 string `json:"type"`
				URI                  string `json:"uri"`
			} `json:"album"`
			Artists []struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"artists"`
			AvailableMarkets []string `json:"available_markets"`
			DiscNumber       int      `json:"disc_number"`
			DurationMs       int      `json:"duration_ms"`
			Explicit         bool     `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string `json:"href"`
			ID          string `json:"id"`
			IsLocal     bool   `json:"is_local"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
		} `json:"items"`
		Limit    int         `json:"limit"`
		Next     interface{} `json:"next"`
		Offset   int         `json:"offset"`
		Previous interface{} `json:"previous"`
		Total    int         `json:"total"`
	} `json:"tracks"`
}
