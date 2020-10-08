package controllers

import (
	"log"
	"zoove/errors"
	"zoove/platforms"
	"zoove/types"
	"zoove/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/soveran/redisurl"
)

type Jaeger struct {
	Pool *redis.Pool
}

// NewJaeger returns a new jaeger (tsk tsk)
func NewJaeger(pool *redis.Pool) *Jaeger {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			// log.Println(os.Getenv("REDIS_URL"))
			return redisurl.Connect()
		},
	}
	return &Jaeger{Pool: pool}
}

// JaegerHandler is the handler for finding tracks on other platforms from one. Using Jaeger for loss of words lol
func (jaeger *Jaeger) JaegerHandler(ctx *fiber.Ctx) error {
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

	var tracks = [][]types.SingleTrack{}
	deezerTracks := []types.SingleTrack{}
	spotifyTracks := []types.SingleTrack{}

	if extracted.Host == util.HostDeezer {
		track, err := platforms.HostDeezerGetSingleTrack(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Println("Error getting the track from Deezer")
			log.Println(err)
			if err == errors.NotFound {
				log.Println("Track does not exist on deezer")
				return util.NotFound(ctx)
			}
		}
		search := platforms.NewTrackToSearch(track.Title, track.Artistes[0], jaeger.Pool)
		spot, err := search.HostSpotifySearchTrack()
		if err != nil {
			if err == errors.NotFound {
				spot = nil
			}
			log.Println("Error fetching spotify search")
			log.Println(err)
		}
		track.ReleaseDate = spot.ReleaseDate
		deezerTracks = append(deezerTracks, *track)
		spotifyTracks = append(spotifyTracks, *spot)
		// tracks = append(tracks, track, spot)
	} else if extracted.Host == util.HostSpotify {
		track, err := platforms.HostSpotifyGetSingleTrack(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Println("Error getting the track from Spotify")
			log.Println(err)
			if err == errors.NotFound {
				log.Println("Track does not exist on spotify")
				return util.NotFound(ctx)
			}
		}
		search := platforms.NewTrackToSearch(track.Title, track.Artistes[0], jaeger.Pool)
		deez, err := search.HostDeezerSearchTrack()
		if err != nil {
			log.Println("Error getting deezer song")
			log.Println(err)
		}
		// this is because spotify always has release date. but now, remember that we're currently
		// checking if track has been cached. We're still leaving this 'cos its caching our calls
		// right now. and thats what we need. but since relasedate real value can be gotten here, simply
		// using the value here.
		deez.ReleaseDate = track.ReleaseDate
		// tracks = append(tracks, track, deez)
		deezerTracks = append(deezerTracks, *deez)
		spotifyTracks = append(spotifyTracks, *track)
	}

	conn := jaeger.Pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("GET", util.RedisSearchesKey))
	if err != nil {
		log.Println("Search counter does not exist.")
		log.Println(err)
		if err == redis.ErrNil {
			_, err := redis.String(conn.Do("SET", util.RedisSearchesKey, "1"))
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

	tracks = append(tracks, deezerTracks, spotifyTracks)
	log.Printf("Searches count is: %d", searchesCount)
	return util.RequestOk(ctx, tracks)
}

// ConvertPlaylist retrieves tracks found for a playlist and returns the equivalents of the playlist tracks from other platforms.
func (jaeger *Jaeger) ConvertPlaylist(ctx *fiber.Ctx) error {
	// lets say someone pasted a spotify URL,
	/**
	FIRST, we want to get the tracks on that playlist.
	Second, then check through cache and see if that song has been cached.
	Third, if cached, then build an array of JSON, each JSON object being songs for the playlist found for other platforms
	**/

	extracted := ctx.Locals("extractedInfo").(*types.ExtractedInfo)
	// log.Printf("Extracted issues: %#v", extracted)

	playlist := &types.Playlist{}
	if extracted.Host == util.HostDeezer {
		deezerPlaylist, err := platforms.HostDeezerFetchPlaylistTracks(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Printf("Error getting user playlist: %s", err.Error())
			util.InternalServerError(ctx, err)
		}
		playlist = &deezerPlaylist
	} else if extracted.Host == util.HostSpotify {
		spotifyPlaylist, err := platforms.HostSpotifyFetchPlaylistTracks(extracted.ID, jaeger.Pool)
		if err != nil {
			log.Printf("Error getting spotify playlist: %s", err.Error())
			return util.InternalServerError(ctx, err)
		}
		// log.Printf("\nFetched playlist is: %#v\n", spotifyPlaylist)
		playlist = &spotifyPlaylist
	}

	outputs := [][]types.SingleTrack{}
	deezerPlaylist := []types.SingleTrack{}
	spotifPlaylist := []types.SingleTrack{}
	spotChan := make(chan *types.SingleTrack)
	deezChan := make(chan *types.SingleTrack)

	for _, singleTrack := range playlist.Tracks {
		search := platforms.NewTrackToSearch(singleTrack.Title, singleTrack.Artistes[0], jaeger.Pool)
		go search.HostDeezerSearchTrackChan(deezChan)
		deezerTrack := <-deezChan
		if deezerTrack == nil {
			continue
		}
		go search.HostSpotifySearchTrackChan(spotChan)
		spotifyTrack := <-spotChan

		if spotifyTrack == nil {
			continue
		}

		deezerPlaylist = append(deezerPlaylist, *deezerTrack)
		spotifPlaylist = append(spotifPlaylist, *spotifyTrack)
	}
	outputs = append(outputs, deezerPlaylist)
	outputs = append(outputs, spotifPlaylist)
	return util.RequestOk(ctx, outputs)
}

// now that we have the playlist for each, we want to look for the equivalent for each track
