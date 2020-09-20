package platforms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"zoove/errors"
	"zoove/types"
	"zoove/util"

	"github.com/gomodule/redigo/redis"
)

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

	err = json.Unmarshal(body, out)
	if err != nil {
		log.Println("Error unserializing the response into json")
		log.Println(err)
		return err
	}
	return nil
}
