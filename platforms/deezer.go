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
	"zoove/errors"
	"zoove/types"
	"zoove/util"

	"github.com/gomodule/redigo/redis"
)

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

func HostDeezerExtractTitle(title string) string {
	// we want to check for the first occurence of "Feat"
	ind := strings.Index(title, "(feat")
	if ind == -1 {
		return title
	}
	out := title[:ind]
	return out
}

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
		key := fmt.Sprintf("%s-%s", util.HostDeezer, string(base.ID))

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
			Cover:    base.Album.Cover,
			Artistes: []string{base.Artist.Name},
			Duration: base.Duration * 1000,
			Explicit: base.ExplicitLyrics,
			ID:       id,
			Platform: util.HostDeezer,
			Preview:  base.Preview,
			Title:    base.Title,
			URL:      base.Link,
			// ReleaseDate: "",
		}

		return track, nil
	}

	return nil, nil
}

func HostDeezerGetSingleTrack(deezerID string, pool *redis.Pool) (*types.SingleTrack, error) {
	conn := pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s-%s", util.HostDeezer, deezerID)
	values, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Println("Error getting from cache")
		log.Println(err)
		if err == redis.ErrNil {
			// TODO: make request to deezer to get here
			log.Println("Error..")
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

func MakeDeezerRequest(url string, out interface{}) error {
	// log.Printf("Deezer UEL is: %s", url)
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

func HostDeezerFetchUserProfile(token string) (*types.HostDeezerUserProfile, error) {
	url := fmt.Sprintf("%s/user/me?access_token=%s", os.Getenv("DEEZER_API_BASE"), token)
	log.Printf("URL is: %s", url)
	profile := &types.HostDeezerUserProfile{}
	err := MakeDeezerRequest(url, profile)
	if err != nil {
		log.Println("Error fetchin user deezer profile")
		log.Println(err)
		return nil, err
	}

	log.Printf("User Deezer profile: %#v", profile)
	return profile, nil
}
