package service

import (
	"context"
	"log"
	"os"
	"zoove/db"
	"zoove/platforms"
	"zoove/types"
	"zoove/util"

	"github.com/fasthttp/websocket"
	"github.com/gomodule/redigo/redis"
	"github.com/soveran/redisurl"
)

var pool *redis.Pool
var jaegerChan = make(chan *SocketMessage)
var spotifyChan = make(chan *types.SingleTrack)
var deezerChan = make(chan *types.SingleTrack)
var createPlaylistChan = make(chan bool)

// SocketMessage represents an incoming socket message
type SocketMessage struct {
	Type    string `json:"action_type"`
	URL     string `json:"url"`
	Payload struct {
		Title    string   `json:"title"`
		Tracks   []string `json:"tracks"`
		Platform string   `json:"platform"`
	} `json:"payload,omitempty"`
	UserID string `json:"userid,omitempty"`
}

// SocketListener represents a "blueprint" for a typical listener
type SocketListener struct {
	Deserialize   SocketMessage
	C             *websocket.Conn
	TrackMeta     *types.SingleTrack
	DeezerTracks  []types.SingleTrack
	SpotifyTracks []types.SingleTrack
	Tracks        [][]types.SingleTrack
	Client        *db.PrismaClient
	PlaylistMeta  *types.Playlist
}

// GetTrackListener listens for tracks action
func (listener *SocketListener) GetTrackListener() {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			log.Println(os.Getenv("REDIS_URL"))
			redisConnect, err := redisurl.Connect()
			if err != nil {
				log.Println("Error with redis connection something something here", err)
			}
			return redisConnect, err
		},
	}

	client := db.NewClient()
	err := client.Connect()

	if err != nil {
		log.Println("Error creating new DB connection")
		log.Fatalln(err)
	}

	defer func() {
		err := client.Disconnect()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	listener.Client = client
	log.Println("DB connection made.")

	// log.Println("Deserialized extracted URL (TRACK) is: ", listener.deserialize.URL)
	extracted, err := util.ExtractInfoMetadata(listener.Deserialize.URL)
	if err != nil {
		log.Println("Error extracting")
		log.Println(err)
		listener.C.WriteMessage(websocket.TextMessage, []byte(`{"desc":"error", "message":"Its me not you...."`))
		listener.C.Close()
	}
	if extracted.Host == util.HostDeezer {
		// log.Println("Wants to search deezer")
		listener.TrackMeta, err = platforms.HostDeezerGetSingleTrack(extracted.ID, pool)
		if err != nil {
			listener.C.WriteMessage(websocket.TextMessage, []byte(`{"desc":"Error getting deezer single track"}`))
			listener.C.Close()
		}

	} else if extracted.Host == util.HostSpotify {
		// log.Println("Wants to search spotify")
		listener.TrackMeta, err = platforms.HostSpotifyGetSingleTrack(extracted.ID, pool)
		if err != nil {
			listener.C.WriteMessage(websocket.TextMessage, []byte(`{"desc":"Error getting spotify single track"}`))
			listener.C.Close()
		}
	} else {
		log.Println("Oops! Not a valid host")
		listener.C.WriteMessage(websocket.TextMessage, []byte(`{"desc":"Invalid host"}`))
		listener.C.Close()
		return
	}
	artiste := ""
	if len(listener.TrackMeta.Artistes) > 0 {
		artiste = listener.TrackMeta.Artistes[0]
	}
	search := platforms.NewTrackToSearch(listener.TrackMeta.Title, artiste, pool)
	deezr, err := search.HostDeezerSearchTrack()
	if err != nil {
		log.Println("Error searching deezer")
		// log.Println("Error is: ", err)
		// TODO: try to handle whatever happens here
		deezr = &types.SingleTrack{}
	}

	spot, err := search.HostSpotifySearchTrack()
	if err != nil {
		// log.Println("Errpr searching spotify")
		// TODO: try to handle whatever happens here
		spot = &types.SingleTrack{}
	}
	conn := pool.Get()
	defer conn.Close()

	_, err = redis.String(conn.Do("GET", util.RedisSearchesKey))
	if err != nil {
		if err == redis.ErrNil {
			_, err := redis.String(conn.Do("SET", util.RedisSearchesKey, "1"))
			if err != nil {
				log.Println("Error saving searches key into redis")
			}
		}
	}

	searchesCount, err := redis.Int(conn.Do("INCR", util.RedisSearchesKey))
	if err != nil {
		log.Println("Error incrementing redis key")
	}
	log.Printf("Number of search so far: %d\n", searchesCount)
	deezr.ReleaseDate = spot.ReleaseDate
	listener.DeezerTracks = append(listener.DeezerTracks, *deezr)
	listener.SpotifyTracks = append(listener.SpotifyTracks, *spot)
	listener.Tracks = append(listener.Tracks, listener.SpotifyTracks, listener.DeezerTracks)
	listener.C.WriteJSON(listener.Tracks)

	// we gotta reset those values, else, it'd just keep pushing to the arrays and returning increasing values as the user makes more requests
	// perhaps have @Davidemi to review this for me.
	listener.Tracks = nil
	listener.DeezerTracks = nil
	listener.SpotifyTracks = nil
	listener.C.Close()
}

// GetPlaylistListener returns the playlist listener
func (listener *SocketListener) GetPlaylistListener() {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			log.Println(os.Getenv("REDIS_URL"))
			redisConnect, err := redisurl.Connect()
			if err != nil {
				log.Println("Error with redis connection something something here", err)
			}
			return redisConnect, err
		},
	}
	// log.Println("Deserialized extracted URL (playlist) is: ", listener.deserialize.URL)
	extracted, err := util.ExtractInfoMetadata(listener.Deserialize.URL)
	if err != nil {
		log.Println("Error extracting")
		log.Println(err)
		listener.C.WriteMessage(websocket.TextMessage, []byte(`{"desc":"error", "message":"Its me not you...."`))
		listener.C.Close()
	}

	if extracted.Host == util.HostDeezer {
		deezerPl, err := platforms.HostDeezerFetchPlaylistTracks(extracted.ID, pool)
		if err != nil {
			log.Println("Error fetching playlist tracks.")
			log.Println(err)
			if err.Error() == "Not Found" {
				listener.PlaylistMeta = &types.Playlist{}
			}
		}

		listener.PlaylistMeta = &deezerPl

		for _, singleTrack := range listener.PlaylistMeta.Tracks {
			search := platforms.NewTrackToSearch(singleTrack.Title, singleTrack.Artistes[0], pool)
			log.Printf("TRACKS UNDER THE PLATFORMS: %v\n\n", search)
			log.Printf("COUNT OF ACTIVE POOL CONNECTION IS: %d \n\n", pool.ActiveCount())
			go search.HostSpotifySearchTrackChan(spotifyChan)
			spotifyTrack := <-spotifyChan

			if spotifyTrack == nil {
				continue
			}

			listener.SpotifyTracks = append(listener.SpotifyTracks, *spotifyTrack)
		}

		listener.DeezerTracks = append(listener.DeezerTracks, listener.PlaylistMeta.Tracks...)

	} else if extracted.Host == util.HostSpotify {
		spotifyPl, err := platforms.HostSpotifyFetchPlaylistTracks(extracted.ID, pool)
		if err != nil {
			log.Println("Error fetching spotify playlist tracks.")
		}
		listener.PlaylistMeta = &spotifyPl

		for _, singleTrack := range listener.PlaylistMeta.Tracks {
			artiste := ""
			if len(singleTrack.Artistes) > 0 {
				artiste = singleTrack.Artistes[0]
			}

			search := platforms.NewTrackToSearch(singleTrack.Title, artiste, pool)
			go search.HostDeezerSearchTrackChan(deezerChan)
			deezerTrack := <-deezerChan
			if deezerTrack == nil {
				continue
			}
			listener.DeezerTracks = append(listener.DeezerTracks, *deezerTrack)
		}
		listener.SpotifyTracks = append(listener.SpotifyTracks, listener.PlaylistMeta.Tracks...)
	}

	conn := pool.Get()
	defer conn.Close()

	_, err = redis.String(conn.Do("GET", util.RedisSearchesKey))
	if err != nil {
		if err == redis.ErrNil {
			_, err := redis.String(conn.Do("SET", util.RedisSearchesKey, "1"))
			if err != nil {
				log.Println("Error saving searches key into redis")
			}
		}
	}

	searchesCount, err := redis.Int(conn.Do("INCR", util.RedisSearchesKey))
	if err != nil {
		log.Println("Error incrementing redis key")
	}
	log.Printf("Number of search so far: %d\n", searchesCount)

	diff := 0
	if len(listener.DeezerTracks) > len(listener.SpotifyTracks) {
		diff = len(listener.DeezerTracks) - len(listener.SpotifyTracks)
		listener.DeezerTracks = listener.DeezerTracks[:len(listener.DeezerTracks)-diff]
	} else if len(listener.SpotifyTracks) > len(listener.DeezerTracks) {
		diff = len(listener.SpotifyTracks) - len(listener.SpotifyTracks)
		listener.SpotifyTracks = listener.SpotifyTracks[:len(listener.SpotifyTracks)-diff]
	}

	for index, single := range listener.DeezerTracks {
		single.ReleaseDate = listener.SpotifyTracks[index].ReleaseDate
	}

	log.Println("Final deezer tracks are: ", listener.DeezerTracks)
	listener.Tracks = append(listener.Tracks, listener.DeezerTracks, listener.SpotifyTracks)
	// log.Println("All tracks now are: ", listener.Tracks)
	// log.Println("Plalyist meta is: ", listener.PlaylistMeta)
	res := map[string]interface{}{
		"playlist_title": listener.PlaylistMeta.Title,
		"payload":        listener.Tracks,
		"owner":          listener.PlaylistMeta.Owner,
		"playlist_meta":  listener.PlaylistMeta,
		"platforms": map[string]interface{}{
			"spotify": listener.SpotifyTracks,
			"deezer":  listener.DeezerTracks,
		},
	}

	listener.C.WriteJSON(res)
	listener.DeezerTracks = nil
	listener.SpotifyTracks = nil
	listener.Tracks = nil
	listener.C.Close()
}

// CreatePlaylistListener creates a playlist for a user.
func (listener *SocketListener) CreatePlaylistListener() {

	existing, _ := listener.Client.User.FindFirst(db.User.UUID.Equals(listener.Deserialize.UserID)).Exec(context.Background())
	log.Printf("\n\nEXISTING USER GOTTEN IS: %v\n\n", existing)
	log.Printf("\n\nPlatform is: %v\n\n", listener.Deserialize.Payload)
	userID := ""
	if listener.Deserialize.Payload.Platform == "spotify" {
		userID = existing.SpotifyID
	} else {
		userID = existing.DeezerID
	}

	go platforms.CreatePlaylistChan(userID, listener.Deserialize.Payload.Title,
		string(existing.Token), listener.Deserialize.Payload.Platform, listener.Deserialize.Payload.Tracks,
		createPlaylistChan)
	_ = <-createPlaylistChan
	res := map[string]interface{}{
		"action":  "create",
		"payload": true,
	}

	listener.C.WriteJSON(res)
	listener.C.Close()
}
