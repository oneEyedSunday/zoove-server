package platforms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"zoove/errors"
	"zoove/types"
	"zoove/util"

	"github.com/gomodule/redigo/redis"
	"github.com/soveran/redisurl"
)

// HostDeezerUserAuth authorizes the user and returns the deezer permanent access_token
func HostDeezerUserAuth(authcode string) (string, error) {
	type deezerToken struct {
		AccessToken string `json:"access_token"`
		Expires     int    `json:"expires"`
	}

	tok := &deezerToken{}
	url := fmt.Sprintf("%s/access_token.php?app_id=%s&secret=%s&code=%s&output=json", os.Getenv("DEEZER_AUTH_BASE"), os.Getenv("DEEZER_APP_ID"), os.Getenv("DEEZER_APP_SECRET"), authcode)
	err := MakeDeezerRequest(url, tok)
	if err != nil {
		log.Println("Error authing user with code.")
		log.Println(err)
		return "", err
	}
	return tok.AccessToken, nil
}

// HostDeezerExtractTitle exteacts and returns the title. It removes (feat <bla bla>) from the title. This is because this title is used to search spotify
func HostDeezerExtractTitle(title string) string {
	// we want to check for the first occurence of "Feat"
	ind := strings.Index(title, "(feat")
	if ind == -1 {
		return title
	}
	out := title[:ind]
	return out
}

// HostDeezerSearchTrackChan searches deezer for a track and returns a single track but using channels
func (search *TrackToSearch) HostDeezerSearchTrackChan(ch chan *types.SingleTrack) {
	conn := search.Pool.Get()
	defer conn.Close()

	title := HostDeezerExtractTitle(search.Title)
	payload := url.QueryEscape(fmt.Sprintf("track:\"%s\" artist:\"%s\"", title, search.Artiste))
	url := fmt.Sprintf("%s/search?q=%s", os.Getenv("DEEZER_API_BASE"), payload)
	output := &types.HostDeezerSearchTrack{}
	err := MakeDeezerRequest(url, output)
	if err != nil {
		log.Println("Error searching on deezer for track")
		log.Println(err)
	}

	// log.Printf("Output from deezer search%#v", output)
	if len(output.Data) > 0 {
		base := output.Data[0]
		key := fmt.Sprintf("%s-%s", util.HostDeezer, strconv.Itoa(base.ID))

		if err != nil {
			log.Println("Error getting from redis")
			log.Println(err)
			if err == redis.ErrNil {
				log.Println("The track has not been previously cached.")
				_, err := redis.String(conn.Do("SET", key, output))
				if err != nil {
					log.Println("Error inserting track into DB")
				}
			}
		}

		id := strconv.Itoa(base.ID)
		track := &types.SingleTrack{
			Cover:       base.Album.Cover,
			Artistes:    []string{base.Artist.Name},
			Duration:    base.Duration * 1000,
			Explicit:    base.ExplicitLyrics,
			ID:          id,
			Platform:    util.HostDeezer,
			Preview:     base.Preview,
			Title:       base.Title,
			URL:         base.Link,
			ReleaseDate: "",
			Album:       base.Album.Title,
		}

		// return track, nil
		ch <- track
		// log.Printf("Track title for song under playlist is: %s\n\n", track.Title)
		return
	}
	// log.Printf("Deezer error: %#v\n\n", err)
	// return nil, errors.NotFound
	ch <- nil
}

// HostDeezerSearchTrack searches deezer for a track and returns a single track
func (search *TrackToSearch) HostDeezerSearchTrack() (*types.SingleTrack, error) {
	conn := search.Pool.Get()
	defer conn.Close()

	title := HostDeezerExtractTitle(search.Title)
	payload := url.QueryEscape(fmt.Sprintf("track:\"%s\" artist:\"%s\"", title, search.Artiste))
	url := fmt.Sprintf("%s/search?q=%s", os.Getenv("DEEZER_API_BASE"), payload)
	output := &types.HostDeezerSearchTrack{}
	err := MakeDeezerRequest(url, output)
	if err != nil {
		log.Println("Error searching on deezer for track")
		log.Println(err)
	}

	// log.Printf("Output from deezer search%#v", output)
	if len(output.Data) > 0 {
		base := output.Data[0]
		key := fmt.Sprintf("%s-%s", util.HostDeezer, strconv.Itoa(base.ID))
		log.Println("Result to search ", base)
		if err != nil {
			log.Println("Error getting from redis")
			log.Println(err)
			if err == redis.ErrNil {
				log.Println("The track has not been previously cached.")
				_, err := redis.String(conn.Do("SET", key, output))
				if err != nil {
					log.Println("Error inserting track into DB")
				}
			}
		}

		id := strconv.Itoa(base.ID)
		log.Println("Single ")
		track := &types.SingleTrack{
			Cover:       base.Album.Cover,
			Artistes:    []string{base.Artist.Name},
			Duration:    base.Duration * 1000,
			Explicit:    base.ExplicitLyrics,
			ID:          id,
			Platform:    util.HostDeezer,
			Preview:     base.Preview,
			Title:       base.Title,
			URL:         base.Link,
			ReleaseDate: "",
			Album:       base.Album.Title,
		}

		return track, nil
	}

	return nil, errors.NotFound
}

// HostDeezerGetSingleTrackChan returns a single deezer track (DOING THE CACHING) but using a go routine
func HostDeezerGetSingleTrackChan(deezerID string, pool *redis.Pool, ch chan *types.SingleTrack) {
	conn := pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s-%s", util.HostDeezer, deezerID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// log.Println("Error getting from cache")
		log.Println(err)
		if err == redis.ErrNil {
			url := fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
			dz := &types.HostDeezerTrack{}
			err = MakeDeezerRequest(url, dz)
			id := strconv.Itoa(dz.ID)
			single := &types.SingleTrack{Cover: dz.Album.Cover, Duration: dz.Duration * 1000, Explicit: dz.ExplicitLyrics, Platform: util.HostDeezer, Preview: dz.Preview, ReleaseDate: dz.ReleaseDate, Title: dz.Title, URL: dz.Link, ID: id}
			for _, elem := range dz.Contributors {
				single.Artistes = append(single.Artistes, elem.Name)
			}

			serialized, err := json.Marshal(single)
			if err != nil {
				log.Println("Error unserializing")
				log.Println(err)
			}

			_, err = redis.String(conn.Do("SET", key, string(serialized)))
			if err != nil {
				log.Println("Error saving into redis.")
				log.Println(err)
			}
			// return single, nil
		}
	}

	single := &types.SingleTrack{}

	err = json.Unmarshal([]byte(values), single)
	if err != nil {
		log.Println("Error serializing the result from redis into a response")
		// return nil, err
		ch <- nil
	}
	ch <- single
	// return single, err
}

// HostDeezerGetSingleTrack returns a single deezer track (DOING THE CACHING)
func HostDeezerGetSingleTrack(deezerID string, pool *redis.Pool) (*types.SingleTrack, error) {
	// pool = &redis.Pool{
	// 	Dial: func() (redis.Conn, error) {
	// 		log.Println(os.Getenv("REDIS_URL"))
	// 		redisConnect, err := redisurl.Connect()
	// 		if err != nil {
	// 			log.Println("Error with redis connection something something here", err)
	// 		}
	// 		return redisConnect, err
	// 	},
	// }

	conn := pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s-%s", util.HostDeezer, deezerID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		// log.Println("Error getting from cache")
		log.Println(err)
		if err == redis.ErrNil {

			url := fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
			dz := &types.HostDeezerTrack{}
			err = MakeDeezerRequest(url, dz)
			id := strconv.Itoa(dz.ID)
			single := &types.SingleTrack{Cover: dz.Album.Cover, Duration: dz.Duration * 1000, Explicit: dz.ExplicitLyrics, Platform: util.HostDeezer, Preview: dz.Preview, ReleaseDate: dz.ReleaseDate, Title: dz.Title, URL: dz.Link, ID: id}
			for _, elem := range dz.Contributors {
				single.Artistes = append(single.Artistes, elem.Name)
			}

			serialized, err := json.Marshal(single)
			if err != nil {
				log.Println("Error unserializing")
				log.Println(err)
			}

			_, err = redis.String(conn.Do("SET", key, string(serialized)))
			if err != nil {
				log.Println("Error saving into redis.")
				log.Println(err)
			}
			return single, nil
		}
	}

	single := &types.SingleTrack{}

	err = json.Unmarshal([]byte(values), single)
	if err != nil {
		log.Println("Error serializing the result from redis into a response")
		return nil, err
	}
	return single, err
}

// MakeDeezerRequest makes an http request to deezer
func MakeDeezerRequest(url string, out interface{}) error {
	log.Printf("Deezer UEL is: %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("Error creating new HTTP request")
		log.Println(err)
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error making request to HTTP")
		log.Println(err)
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	// log.Printf("Body of response: %s", string(body))
	if strings.Contains(string(body), "{\"error\"") {
		return errors.NotFound
	}
	if err != nil {
		log.Println("Error reading response body into memory")
		return err
	}
	if res.StatusCode == http.StatusUnauthorized {
		return errors.UnAuthorized
	}
	if res.StatusCode == http.StatusInternalServerError {
		return err
	}

	// log.Printf("Body is: %s", string(body))
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Println("Error unserializing the response into json")
		log.Println(err)
		return err
	}
	return nil
}

// HostDeezerFetchUserProfile returns a user's profile and an error if any.
func HostDeezerFetchUserProfile(token string) (*types.HostDeezerRawUserProfile, error) {
	url := fmt.Sprintf("%s/user/me?access_token=%s", os.Getenv("DEEZER_API_BASE"), token)
	profile := &types.HostDeezerRawUserProfile{}
	err := MakeDeezerRequest(url, profile)
	if err != nil {
		log.Println("Error fetchin user deezer profile")
		log.Println(err)
		return nil, err
	}

	return profile, nil
}

// HostDeezerFetchHistory returns an array of the tracks a user recently played
func HostDeezerFetchHistory(token string) ([]types.SingleTrack, error) {

	url := fmt.Sprintf("%s/user/me/history?access_token=%s", os.Getenv("DEEZER_API_BASE"), token)
	history := &types.HostDeezerHistory{}
	err := MakeDeezerRequest(url, history)
	if err != nil {
		log.Println("Error fetching history")
		return nil, err
	}

	histr := []types.SingleTrack{}
	for _, track := range history.Data {
		id := strconv.Itoa(track.ID)
		tm := time.Unix(int64(track.Timestamp), 0)
		single := types.SingleTrack{
			Cover:       track.Album.Cover,
			Duration:    track.Duration * 1000, // to ms
			Explicit:    track.ExplicitLyrics,
			ID:          id,
			Platform:    util.HostDeezer,
			Preview:     track.Preview,
			ReleaseDate: "",
			PlayedAt:    tm.String(),
			Title:       track.Title,
			URL:         track.Link,
			Artistes:    []string{track.Artist.Name},
		}
		histr = append(histr, single)
	}
	return histr, nil
}

// HostDeezerFetchArtisteHistory returns the artistes listening history of a user.
func HostDeezerFetchArtisteHistory(token string) ([]string, error) {
	hist, err := HostDeezerFetchHistory(token)
	if err != nil {
		return nil, err
	}

	history := []string{}
	for _, track := range hist {
		history = append(history, track.Artistes...)
	}
	return history, nil
}

// HostDeezerFetchPlaylistTracks returns the deezer playlist information
func HostDeezerFetchPlaylistTracks(playlistID string, pool *redis.Pool) (types.Playlist, error) {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			// log.Println(os.Getenv("REDIS_URL"))
			return redisurl.Connect()
		},
	}
	conn := pool.Get()
	defer conn.Close()

	deezerPlaylist := &types.HostDeezerPlaylistResponse{}

	deezerBaseAPI := os.Getenv("DEEZER_API_BASE")
	url := fmt.Sprintf("%s/playlist/%s", deezerBaseAPI, playlistID)

	// log.Println("URL to get deezer playlist is: ", url)
	err := MakeDeezerRequest(url, deezerPlaylist)
	if err != nil {
		return types.Playlist{}, err
	}
	id := strconv.Itoa(deezerPlaylist.ID)
	playlist := &types.Playlist{Description: deezerPlaylist.Description, Collaborative: deezerPlaylist.Collaborative,
		Duration: deezerPlaylist.Duration, TracksNumber: deezerPlaylist.NbTracks, Title: deezerPlaylist.Title,
		Tracks: []types.SingleTrack{}, Owner: types.PlaylistOwner{Avatar: deezerPlaylist.Picture, ID: id, Name: deezerPlaylist.Creator.Name},
		URL: deezerPlaylist.Link, Cover: deezerPlaylist.Picture,
	}
	for _, track := range deezerPlaylist.Tracks.Data {
		dur := strconv.Itoa(track.Duration)
		trackDuration, _ := strconv.ParseInt(dur, 10, 64)
		timeAdded := time.Unix(int64(track.TimeAdd), 0)
		tID := strconv.Itoa(track.ID)
		single := &types.SingleTrack{Cover: track.Album.Cover, Artistes: []string{track.Artist.Name}, Duration: int(trackDuration) * 1000, Explicit: track.ExplicitLyrics,
			ID: tID, Platform: util.HostDeezer, Preview: track.Preview, Title: track.Title, AddedAt: timeAdded.String(),
			URL: track.Link,
		}
		playlist.Tracks = append(playlist.Tracks, *single)
	}

	if err != nil {
		return types.Playlist{}, err
	}
	// log.Printf("Playlist is: %#v", playlist)
	return *playlist, nil
}

// HostDeezerCreatePlaylist creates a new playlist for the deezer user
func HostDeezerCreatePlaylist(title, userid, token string, tracks []string) error {
	deezerAPIBase := os.Getenv("DEEZER_API_BASE")
	url := fmt.Sprintf("%s/user/%s/playlists?access_token=%s&request_method=post&title=%s", deezerAPIBase, userid, token, title)
	src := &types.DeezerPlaylistCreationResponse{}
	err := util.MakeRequest(url, src)
	if err != nil {
		log.Println("Error making request here.")
		log.Println(err)
		return err
	}

	allTracks := strings.Join(tracks, ",")
	playlistURL := fmt.Sprintf("%s/playlist/%d/tracks?access_token=%s&request_method=post&songs=%s", deezerAPIBase, src.ID, token, allTracks)
	err = util.MakeRequest(playlistURL, true)

	if err != nil {
		log.Println("Error  making request to add tracks to playlist")
		log.Println(err)
		return err
	}
	return nil
}
