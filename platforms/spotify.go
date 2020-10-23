package platforms

import (
	"bytes"
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
	"github.com/soveran/redisurl"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var scopes = url.QueryEscape(fmt.Sprintf("%s %s %s %s %s %s %s", spotify.ScopeUserReadPrivate, spotify.ScopeUserReadEmail,
	spotify.ScopePlaylistModifyPublic, spotify.ScopeUserLibraryModify,
	spotify.ScopeUserTopRead, spotify.ScopeUserReadRecentlyPlayed,
	spotify.ScopeUserReadCurrentlyPlaying))

// HostSpotifySearchTrackChan returns a searched track using channels
func (search *TrackToSearch) HostSpotifySearchTrackChan(ch chan *types.SingleTrack) {
	payload := url.QueryEscape(fmt.Sprintf("track:%s artist:%s", search.Title, search.Artiste))
	searchURL := fmt.Sprintf("%s/v1/search?q=%s&type=track", os.Getenv("SPOTIFY_API_BASE"), payload)
	output := &types.HostSpotifySearchTrack{}
	token, err := GetSpotifyAuthToken()
	if err != nil {
		// return nil, err
		ch <- nil
	}

	err = MakeSpotifyRequest(searchURL, token.AccessToken, output)
	if err != nil {
		// return nil, err
		ch <- nil
	}
	// log.Printf("\nOUTPUT HERE: %#v\n\n", output.Tracks)
	if len(output.Tracks.Items) > 0 {

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
			// return track, nil
			ch <- track
			// log.Printf("TITLE OF TRACK ON SPOTIFY IS: %s", track.Title)
			return
		}
	}
	// return nil, errors.NotFound
	ch <- nil
}

// HostSpotifySearchTrack returns a searched track.
func (search *TrackToSearch) HostSpotifySearchTrack() (*types.SingleTrack, error) {
	payload := url.QueryEscape(fmt.Sprintf("track:%s artist:%s", search.Title, search.Artiste))
	searchURL := fmt.Sprintf("%s/v1/search?q=%s&type=track", os.Getenv("SPOTIFY_API_BASE"), payload)
	output := &types.HostSpotifySearchTrack{}
	token, err := GetSpotifyAuthToken()
	if err != nil {
		return nil, err
	}

	err = MakeSpotifyRequest(searchURL, token.AccessToken, output)
	if err != nil {
		return nil, err
	}
	// log.Printf("\nOUTPUT HERE: %#v\n\n", output.Tracks)
	if len(output.Tracks.Items) > 0 {

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
	}
	return nil, errors.NotFound
}

// HostSpotifyReturnAuth returns a new oauth token for spotify user. Note this is not used used for making calls that require user permission
func HostSpotifyReturnAuth(authcode string) (*oauth2.Token, error) {
	spotifyAuthBaseURL := os.Getenv("SPOTIFY_AUTH_BASE")
	spotifyRedirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	spotifyBearer := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifyClientSecret))

	reqbody := url.Values{}
	reqbody.Set("grant_type", "authorization_code")
	reqbody.Set("code", authcode)
	reqbody.Set("redirect_uri", spotifyRedirectURI)

	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/api/token", spotifyAuthBaseURL)
	r, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqbody.Encode()))

	r.Header.Set("Authorization", "Basic "+spotifyBearer)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Content-Length", strconv.Itoa(len(reqbody.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := &oauth2.Token{}
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, types.UnAuthorizedScope
	}

	err = json.Unmarshal(body, res)
	if err != nil {
		log.Fatalln(err)
	}
	// log.Println("Response of oauth should be: ", res.RefreshToken)

	// log.Println(string(body))
	return res, nil
}

// HostSpotifyUserAuth authorizes a user and returns the spotify user profile
func HostSpotifyUserAuth(authcode string) (*spotify.PrivateUser, string, error) {
	redirecURI := os.Getenv("spotifyRedirectURI")
	token, err := HostSpotifyReturnAuth(authcode)
	if err != nil {
		return nil, "", err
	}

	auth := spotify.NewAuthenticator(redirecURI, spotify.ScopeUserReadPrivate, spotify.ScopeUserReadEmail,
		spotify.ScopePlaylistModifyPublic, spotify.ScopeUserLibraryModify,
		spotify.ScopeUserTopRead, spotify.ScopeUserReadRecentlyPlayed,
		spotify.ScopeUserReadCurrentlyPlaying,
	)

	auth.SetAuthInfo(os.Getenv("spotifyClientID"), os.Getenv("spotifyClientSecret"))
	client := auth.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, "", err
	}

	return user, token.RefreshToken, nil
}

// HostSpotifyGetSingleTrackChan returns a single (cached) spotify track but using a channel
func HostSpotifyGetSingleTrackChan(spotifyID string, pool *redis.Pool, ch chan *types.SingleTrack) {
	conn := pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("%s-%s", "spotify", spotifyID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// log.Println("Error getting single track")
		if err == redis.ErrNil {
			tokens, err := GetSpotifyAuthToken()
			if err != nil {
				// return nil, err
				ch <- nil
			}

			sptf := &types.HostSpotifyTrack{}
			err = MakeSpotifyRequest(fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("spotifyApiBase"), spotifyID), tokens.AccessToken, sptf)

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
				// return nil, err
				ch <- nil
			}
			_, err = redis.String(conn.Do("SET", key, string(serialize)))
			if err != nil {
				// just log. not handling this error as its none crucial. users dont care it doesnt impact them
				log.Println("Error inserting into redis")
				log.Println(err)
			}
			// return single, err
			ch <- single
			return
		}
	}

	single := &types.SingleTrack{}
	err = json.Unmarshal([]byte(values), single)
	if err != nil {
		// return nil, err
		ch <- nil
		return
	}

	ch <- single
}

// HostSpotifyGetSingleTrack returns a single (cached) spotify track
func HostSpotifyGetSingleTrack(spotifyID string, pool *redis.Pool) (*types.SingleTrack, error) {
	conn := pool.Get()
	defer conn.Close()
	key := fmt.Sprintf("%s-%s", "spotify", spotifyID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// log.Println("Error getting single track")
		if err == redis.ErrNil {
			tokens, err := GetSpotifyAuthToken()
			if err != nil {
				return nil, err
			}

			sptf := &types.HostSpotifyTrack{}
			err = MakeSpotifyRequest(fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), spotifyID), tokens.AccessToken, sptf)
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
				return nil, err
			}
			_, err = redis.String(conn.Do("SET", key, string(serialize)))
			if err != nil {
				// just log. not handling this error as its none crucial. users dont care it doesnt impact them
				log.Println("Error inserting into redis")
				log.Println(err)
			}
			return single, err
		}
	}

	single := &types.SingleTrack{}
	err = json.Unmarshal([]byte(values), single)
	if err != nil {
		return nil, err
	}

	return single, nil
}

// HostSpotifyListeningHistory returns the listening history for a spotify user
func HostSpotifyListeningHistory(refreshToken string) ([]types.SingleTrack, error) {
	spotifyAPIBase := os.Getenv("SPOTIFY_API_BASE")
	accessToken, err := HostSpotifyGetAuthorizedAcessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/me/player/recently-played", spotifyAPIBase)
	history := &types.HostSpotifyHistory{}
	err = MakeSpotifyRequest(url, accessToken.AccessToken, history)
	if err != nil {
		// log.Printf("Error making request to the listening history: %s", err)
		return nil, err
	}

	hist := []types.SingleTrack{}
	for _, h := range history.Items {
		base := h.Track.Album
		img := ""
		if len(base.Images) > 0 {
			img = base.Images[0].URL
		}
		artistes := []string{}
		for _, k := range h.Track.Artists {
			artistes = append(artistes, k.Name)
		}

		track := types.SingleTrack{Cover: img, Artistes: artistes, Duration: h.Track.DurationMs,
			Explicit: h.Track.Explicit, ID: h.Track.ID, Platform: util.HostSpotify, Preview: h.Track.PreviewURL,
			ReleaseDate: h.Track.Album.ReleaseDate, Title: h.Track.Name, URL: h.Track.ExternalUrls.Spotify, PlayedAt: h.PlayedAt.String(),
		}
		hist = append(hist, track)
	}

	return hist, nil
}

// HostSpotifyGetAuthorizedAcessToken returns a user authorized token. this is different from GetSpotifyAuthToken because this one can be
// used for user authorization required actions (for example, getting play history).
// Use this only when you need to make calls that require user access. this is because it has lower rate limit.
func HostSpotifyGetAuthorizedAcessToken(refreshToken string) (*types.HostSpotifyAccessTokenRefreshResponse, error) {
	spotifyAuthBaseURL := os.Getenv("SPOTIFY_AUTH_BASE")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyBearer := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifyClientSecret))

	reqBody := url.Values{}
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/token", spotifyAuthBaseURL)
	reqBody.Set("grant_type", "refresh_token")
	reqBody.Set("refresh_token", refreshToken)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+spotifyBearer)
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
	authRes := &types.HostSpotifyAccessTokenRefreshResponse{}
	err = json.Unmarshal(body, authRes)
	if err != nil {
		return nil, err
	}
	return authRes, nil
}

// GetSpotifyAuthToken returns a normal spotify oauth token for a us. this token is used for things that dont require user permission or scopes
func GetSpotifyAuthToken() (*oauth2.Token, error) {
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyBearer := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifyClientSecret))

	spotifyAuthBaseURL := os.Getenv("SPOTIFY_AUTH_BASE")
	reqBody := url.Values{}
	reqBody.Set("grant_type", "client_credentials")

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/token", spotifyAuthBaseURL)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(reqBody.Encode()))
	if err != nil {
		log.Fatalf("Error with spotify auth")
	}

	req.Header.Set("Authorization", "Basic "+spotifyBearer)
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
	out := &oauth2.Token{}

	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}

	return out, nil

}

// HostSpotifyFetchArtisteHistory returns the artistes user has listened to recently
func HostSpotifyFetchArtisteHistory(token string) ([]string, error) {
	hist, err := HostSpotifyListeningHistory(token)
	if err != nil {
		return nil, err
	}

	history := []string{}
	for _, track := range hist {
		history = append(history, track.Artistes...)
	}
	return history, nil
}

// HostSpotifyFetchPlaylistTracks returns a cached spotify playlist
func HostSpotifyFetchPlaylistTracks(playlistID string, pool *redis.Pool) (types.Playlist, error) {
	// log.Printf("PLAYLIST IS %s\n", playlistID)
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			// log.Println(os.Getenv("REDIS_URL"))
			return redisurl.Connect()
		},
	}

	tok, err := GetSpotifyAuthToken()
	if err != nil {
		return types.Playlist{}, err
	}
	// log.Printf("\nReturned token: %#v", tok.AccessToken)

	conn := pool.Get()
	defer conn.Close()

	auth := spotify.NewAuthenticator(os.Getenv("SPOTIFY_REDIRECT_URI"), scopes)
	client := auth.NewClient(tok)
	spotifyPlaylist, err := client.GetPlaylist(spotify.ID(playlistID))
	if err != nil {
		return types.Playlist{}, err
	}
	durationMs := 0
	avatar := ""
	if len(spotifyPlaylist.Owner.Images) > 0 {
		avatar = spotifyPlaylist.Owner.Images[0].URL
	}
	playlist := types.Playlist{Description: spotifyPlaylist.Description,
		Collaborative: spotifyPlaylist.Collaborative,
		Title:         spotifyPlaylist.Name,
		Owner: types.PlaylistOwner{Avatar: avatar, ID: spotifyPlaylist.Owner.ID,
			Name: spotifyPlaylist.Name},
	}
	for _, single := range spotifyPlaylist.Tracks.Tracks {
		durationMs += single.Track.Duration
		singleT := &types.SingleTrack{
			AddedAt:     single.AddedAt,
			Cover:       single.Track.Album.Images[0].URL,
			Duration:    single.Track.Duration,
			Explicit:    single.Track.Explicit,
			ID:          single.Track.ID.String(),
			Platform:    util.HostSpotify,
			Title:       single.Track.Name,
			URL:         single.Track.Endpoint,
			ReleaseDate: single.Track.Album.ReleaseDate,
			Preview:     single.Track.PreviewURL,
		}
		for _, r := range single.Track.Artists {
			singleT.Artistes = append(singleT.Artistes, r.Name)
		}
		playlist.Tracks = append(playlist.Tracks, *singleT)
	}
	playlist.Duration = durationMs

	if err != nil {
		return types.Playlist{}, nil
	}
	if err != nil {
		return types.Playlist{}, nil
	}

	// log.Println("Tracks found for the playlist is: ", playlist)
	return playlist, nil
}

// HostSpotifyCreatePlaylist creates a playlist with tracks for a user
func HostSpotifyCreatePlaylist(spotifyID, title, token string, tracks []string) error {
	spotifyAPIBase := os.Getenv("SPOTIFY_API_BASE")

	url := fmt.Sprintf("%s/v1/users/%s/playlists", spotifyAPIBase, spotifyID)
	createdPlaylist := &types.HostSpotifyNewPlaylistCreationResponse{}
	playlist := types.HostSpotifyCreatePlaylist{Name: title}
	bodyJSON, err := json.Marshal(playlist)
	if err != nil {
		log.Println("Error serializing playlist creation JSON")
		log.Println(err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		log.Println("Error POSTing to endpoint")
		log.Println(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	err = ExecuteRequest(req, createdPlaylist)

	if err != nil {
		log.Println("Error making spotify post request to create playlist")
		log.Println(err)
		return err
	}

	var spotifyURIs []string
	for _, elem := range tracks {
		formatted := fmt.Sprintf("spotify:track:%s", elem)
		spotifyURIs = append(spotifyURIs, formatted)
	}

	url = fmt.Sprintf("%s/v1/playlists/%s/tracks?uris=%s", spotifyAPIBase, createdPlaylist.ID, strings.Join(spotifyURIs, ","))
	log.Println("URL of the songs to add: ", url)
	spotifyPlaylist := &types.HostSpotifyAddNewPlaylistTracksResponse{}
	req, err = http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println("Error GETing")
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	err = ExecuteRequest(req, spotifyPlaylist)
	if err != nil {
		return err
	}

	return nil
}

// ExecuteRequest executes an http API call and deserializes the returned data into an input result
func ExecuteRequest(req *http.Request, result interface{}) error {
	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println("Error executing request call")
		log.Println(err)
		return err
	}
	defer response.Body.Close()
	out, err := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusUnauthorized {
		log.Println("DOes not have permission to perform that action")
		return types.UnAuthorizedScope
	}
	err = json.NewDecoder(bytes.NewReader(out)).Decode(result)
	if err != nil {
		log.Println("Error deserializing body in JSON Decoder")
		return err
	}
	return nil
}

// MakeSpotifyRequest makes a spotify API call
func MakeSpotifyRequest(url, token string, out interface{}) error {
	// log.Printf("URL is: %s", url)
	// log.Printf("Token is: %s", token)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
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
