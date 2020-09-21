package platforms

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"zoove/errors"
	"zoove/types"
	"zoove/util"

	"github.com/gomodule/redigo/redis"
)

type TrackToSearch struct {
	Title   string
	Artiste string
	Pool    *redis.Pool
}

func NewTrackToSearch(title, artiste string, pool *redis.Pool) *TrackToSearch {
	return &TrackToSearch{Artiste: artiste, Title: title, Pool: pool}
}

func (search *TrackToSearch) HostSpotifySearchTrack() (*types.SingleTrack, error) {
	payload := url.QueryEscape(fmt.Sprintf("track:%s artist:%s", search.Title, search.Artiste))
	searchURL := fmt.Sprintf("%s/v1/search?q=%s&type=track", os.Getenv("SPOTIFY_API_BASE"), payload)
	output := &types.HostSpotifySearchTrack{}
	token, err := GetSpotifyAuthToken()
	if err != nil {
		log.Println("Error authenticating spotify and returning needed tokens")
		log.Println(err)
	}

	err = MakeSpotifyRequest(searchURL, token.AccessToken, output)
	if err != nil {
		log.Println("Error authorizing spotify")
		log.Println(err)
		return nil, err
	}

	if len(output.Tracks.Items[0].Artists) > 0 {
		base := output.Tracks.Items[0]
		artistes := []string{}
		for i := range output.Tracks.Items[0].Artists {
			artistes = append(artistes, output.Tracks.Items[0].Artists[i].Name)
		}
		track := &types.SingleTrack{
			Cover:       base.Album.Images[0].URL,
			Duration:    base.DurationMs,
			Explicit:    base.Explicit,
			ID:          base.ID,
			Platform:    util.HostSpotify,
			Preview:     base.PreviewURL,
			ReleaseDate: base.Album.ReleaseDate,
			Title:       base.Name,
			URL:         base.ExternalUrls.Spotify,
			Artistes:    artistes,
		}
		return track, nil
	}
	return nil, nil
}

func HostSpotifyGetSingleTrack(spotifyID string, pool *redis.Pool) (*types.SingleTrack, error) {
	conn := pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("%s-%s", "spotify", spotifyID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Println("Error getting single track")
		if err == redis.ErrNil {
			// payload := fmt.Sprintf("track:%s artist:%s")
			// escaped := url.QueryEscape(payload)
			// token := &types.HostSpotifyAuthResponse{}
			tokens, err := GetSpotifyAuthToken()
			log.Printf("Spotify auth response is: %#v", tokens)
			if err != nil {
				log.Println("Error getting the spotify token")
				log.Println(err)
				return nil, err
			}

			sptf := &types.HostSpotifyTrack{}
			err = MakeSpotifyRequest(fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), spotifyID), tokens.AccessToken, sptf)
			log.Printf("ody")
			log.Printf("SPOTIFY SEARCH IS: %#v", sptf)
			single := &types.SingleTrack{
				Cover:       sptf.Album.Images[0].URL,
				Duration:    sptf.DurationMs,
				Explicit:    sptf.Explicit,
				ID:          sptf.ID,
				Platform:    util.HostSpotify,
				Preview:     sptf.PreviewURL,
				ReleaseDate: sptf.Album.ReleaseDate,
				Title:       sptf.Name,
				URL:         sptf.ExternalUrls.Spotify,
			}
			for _, elem := range sptf.Artists {
				single.Artistes = append(single.Artistes, elem.Name)
			}

			serialize, err := json.Marshal(single)
			if err != nil {
				log.Println("Error serializing for saving into the DB")
				return nil, err
			}
			_, err = redis.String(conn.Do("SET", key, string(serialize)))
			if err != nil {
				log.Println("Error inserting into redis")
			}
			return single, err
		}
	}

	single := &types.SingleTrack{}
	err = json.Unmarshal([]byte(values), single)
	if err != nil {
		log.Println("Error deserializing the cached value")
		log.Println(err)
		return nil, err
	}

	return single, nil
}

func GetSpotifyAuthToken() (*types.HostSpotifyAuthResponse, error) {

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifySecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/token", os.Getenv("SPOTIFY_AUTH_BASE"))
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody.Encode()))
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
	out := &types.HostSpotifyAuthResponse{}

	err = json.Unmarshal(body, out)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return out, nil

}

func MakeSpotifyRequest(url, token string, out interface{}) error {
	// log.Printf("URL is: %s", url)
	// log.Printf("Token is: %s", token)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		log.Println("Error making request to the URL")
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
		return errors.UnAuthorized
	} else if res.StatusCode == http.StatusNotFound {
		return errors.NotFound
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Println("Error deserializing body into JSON")
		return err
	}
	return nil
}
