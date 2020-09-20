package controllers

import (
	"log"
	"zoove/errors"
	"zoove/platforms"
	"zoove/types"
	"zoove/util"

	"github.com/gofiber/fiber"
	"github.com/gomodule/redigo/redis"
)

type Jaeger struct {
	Pool *redis.Pool
}

// NewJaeger returns a new jaeger (tsk tsk)
func NewJaeger(pool *redis.Pool) *Jaeger {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}
	return &Jaeger{Pool: pool}
}

// JaegerHandler is the handler for finding tracks on other platforms from one. Using Jaeger for loss of words lol
func (jaeger *Jaeger) JaegerHandler(ctx *fiber.Ctx) {
	extracted := ctx.Locals("extractedInfo").(*types.ExtractedInfo)

	// lets say that user pastes deezer link. The extractedInfo would be:
	/**
	{
		Host: "deezer",
		URL: "https://api.deezer.com/tracks/453948934",
		ID: 453948934
	}
	now, we make a request to get the information for this song. this would allow us use that information to search on other
	platforms.
	==> Get track info
	==> Pass the track info into (lets say) Spotify
	==> SPotiy also gets the track. (but in this case, it uses search since we dont know the ID of the song on Spotify)
	==> Deezer and Spotify both put their values in an array.
	==> Array is returned.
	==> If Spotify doesnt have it, return 404. Return null (nil) for the Spotify response.
	==> If deezer is empty, return 404 but dont trigger other platforms
	**/

	var tracks = []*types.SingleTrack{}
	if extracted.Host == util.HostDeezer {
		track, err := platforms.HostDeezerGetSingleTrack(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Println("Error getting the track from Deezer")
			log.Println(err)
			if err == errors.NotFound {
				log.Println("Track does not exist on deezer")
				util.NotFound(ctx)
				return
			}
		}
		tracks = append(tracks, track)
	} else if extracted.Host == util.HostSpotify {
		log.Println("Its spotify")
		track, err := platforms.HostSpotifyGetSingleTrack(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Println("Error getting the track from Spotify")
			log.Println(err)
			if err == errors.NotFound {
				log.Println("Track does not exist on spotify")
				util.NotFound(ctx)
				return
			}
		}
		tracks = append(tracks, track)
	}

	conn := jaeger.Pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("GET", util.RedisSearchesKey))
	if err != nil {
		log.Println("Search counter does not exist.")
		log.Println(err)
		if err == redis.ErrNil {
			_, err := redis.String(conn.Do("SET", "seaches", "1"))
			if err != nil {
				log.Println("Error saving key into redis")
				log.Println(err)
			}
		}
	}

	searchesCount, err := redis.Int(conn.Do("INCR", util.RedisSearchesKey))
	if err != nil {
		log.Println("Error incrememnting redis key")
		log.Println(err)
	}

	log.Printf("Searches count is: %d", searchesCount)
	util.RequestOk(ctx, tracks)
	return
}
