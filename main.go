package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"zoove/controllers"
	"zoove/db"
	"zoove/middleware"
	"zoove/platforms"
	"zoove/types"
	"zoove/util"

	"github.com/gofiber/websocket/v2"
	"github.com/soveran/redisurl"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

var pool *redis.Pool
var register = make(chan *websocket.Conn)
var jaegerChan = make(chan *SocketMessage)
var spotifyChan = make(chan *types.SingleTrack)
var deezerChan = make(chan *types.SingleTrack)

func loadEnv() {
	envr := os.Getenv("ENV")
	err := godotenv.Load(".env." + envr)
	if err != nil {
		log.Println("Error reading the env file")
		log.Println(err)
		// panic(err)
	}
}

func init() {
	loadEnv()
}

type SocketMessage struct {
	Type string `json:"action_type"`
	URL  string `json:"url"`
}

func loadListeners() {
	for {
		select {
		case _ = <-register:
			// log.Println("New client connected..")
		}
	}
}

func main() {
	app := fiber.New()

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

	userHandler := controllers.NewUserHandler(client, pool)
	jaeger := controllers.NewJaeger(pool)
	authentication := middleware.NewAuthUserMiddleware(client)

	go loadListeners()

	app.Use(cors.New(cors.Config{
		// AllowOrigins: "*",
		AllowMethods: fmt.Sprintf("%s,%s,%s,%s,%s", http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodOptions, http.MethodDelete),
	}))

	type Sample struct {
		AccessToken string `query:"access_token"`
	}

	app.Get("/deezer/join", func(c *fiber.Ctx) error {
		DeezerAuthBase := os.Getenv("DEEZER_AUTH_BASE")
		DeezerAppID := os.Getenv("DEEZER_APP_ID")
		DeezerRedirectURI := os.Getenv("DEEZER_REDIRECT_URI")
		scopes := "basic_access,email,offline_access,listening_history"
		url := fmt.Sprintf("%s/auth.php?app_id=%s&redirect_uri=%s&perms=%s", DeezerAuthBase, DeezerAppID, DeezerRedirectURI, scopes)
		return c.Redirect(url)
	})

	app.Use("/api/v1.1/ws", func(ctx *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(ctx) {
			ctx.Locals("allowed", true)
			return ctx.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/api/v1.1/ws/connect", websocket.New(func(c *websocket.Conn) {
		var tracks = [][]types.SingleTrack{}
		var deezerTracks = []types.SingleTrack{}
		var spotifyTracks = []types.SingleTrack{}

		register <- c
		for {
			pool = &redis.Pool{
				Dial: func() (redis.Conn, error) {
					// log.Println(os.Getenv("REDIS_URL"))
					return redisurl.Connect()
				},
			}
			_, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("Read Error:", err)
				}
				return
			}
			deserialize := &SocketMessage{}
			err = json.Unmarshal(msg, deserialize)
			if err != nil {
				log.Println("Error parsing. Seems client is sending non-json data")
				c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"send JSON unmarshalling errors here"}`))
				c.Close()
			}

			var trackMeta = &types.SingleTrack{}
			var playlistMeta = &types.Playlist{}

			if deserialize.Type == "track" {

				extracted, err := util.ExtractInfoMetadata(deserialize.URL)
				if err != nil {
					log.Println("Error extracting")
					log.Println(err)
					c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"error", "message":"Its me not you...."`))
					c.Close()
				}
				if extracted.Host == util.HostDeezer {
					// log.Println("Wants to search deezer")
					trackMeta, err = platforms.HostDeezerGetSingleTrack(extracted.ID, pool)
					if err != nil {
						c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"Error getting deezer single track"}`))
						c.Close()
					}

				} else if extracted.Host == util.HostSpotify {
					// log.Println("Wants to search spotify")
					trackMeta, err = platforms.HostSpotifyGetSingleTrack(extracted.ID, pool)
					if err != nil {
						c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"Error getting spotify single track"}`))
						c.Close()
					}
				}

				search := platforms.NewTrackToSearch(trackMeta.Title, trackMeta.Artistes[0], pool)
				deezr, err := search.HostDeezerSearchTrack()
				if err != nil {
					log.Println("Error searching deezer")
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
				deezerTracks = append(deezerTracks, *deezr)
				spotifyTracks = append(spotifyTracks, *spot)
				tracks = append(tracks, spotifyTracks, deezerTracks)
				c.WriteJSON(tracks)

				// we gotta reset those values, else, it'd just keep pushing to the arrays and returning increasing values as the user makes more requests
				// perhaps have @Davidemi to review this for me.
				tracks = nil
				deezerTracks = nil
				spotifyTracks = nil
				c.Close()
			} else {
				extracted, err := util.ExtractInfoMetadata(deserialize.URL)
				if err != nil {
					log.Println("Error extracting")
					log.Println(err)
					c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"error", "message":"Its me not you...."`))
					c.Close()
				}

				if extracted.Host == util.HostDeezer {
					deezerPl, err := platforms.HostDeezerFetchPlaylistTracks(extracted.ID, pool)
					if err != nil {
						log.Println("Error fetching playlist tracks.")
						// TODO: try to handle whatever happens here
					}
					playlistMeta = &deezerPl
				} else if extracted.Host == util.HostSpotify {
					spotifyPl, err := platforms.HostSpotifyFetchPlaylistTracks(extracted.ID, pool)
					if err != nil {
						log.Println("Error fetching spotify playlist tracks.")
					}
					playlistMeta = &spotifyPl
				}

				for _, singleTrack := range playlistMeta.Tracks {
					search := platforms.NewTrackToSearch(singleTrack.Title, singleTrack.Artistes[0], pool)
					go search.HostDeezerSearchTrackChan(deezerChan)
					deezerTrack := <-deezerChan
					if deezerTrack == nil {
						continue
					}
					go search.HostSpotifySearchTrackChan(spotifyChan)
					spotifyTrack := <-spotifyChan
					if spotifyTrack == nil {
						continue
					}

					deezerTracks = append(deezerTracks, *deezerTrack)
					spotifyTracks = append(spotifyTracks, *spotifyTrack)
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

				tracks = append(tracks, deezerTracks, spotifyTracks)
				c.WriteJSON(tracks)
				deezerTracks = nil
				spotifyTracks = nil
				tracks = nil
				c.Close()
			}

		}
	}))

	app.Get("/deezer/verify", userHandler.VerifyDeezerSignup)
	app.Get("/:platform/oauth", userHandler.AuthorizeUser)
	app.Post("/api/v1.1/user/join", userHandler.AddNewUser)
	app.Use(middleware.ExtractedInfoMiddleware)
	app.Get("/api/v1.1/search", jaeger.JaegerHandler)
	app.Get("/api/v1.1/zoovify/playlist", jaeger.ConvertPlaylist)

	app.Use(jwtware.New(
		jwtware.Config{SigningKey: []byte(os.Getenv("JWT_SECRET")),
			Claims:     &types.Token{},
			ContextKey: "user",
		}))
	app.Use(authentication.AuthenticateUser)
	app.Get("/api/v1.1/me", userHandler.GetUserProfile)
	app.Get("/api/v1.1/me/update", userHandler.UpdateUserProfile)
	app.Get("/api/v1.1/me/history", userHandler.GetListeningHistory)
	app.Get("/api/v1.1/me/history/artistes", userHandler.GetArtistePlayHistory)

	// app.Get("/api/v1.1/me/history")

	port := os.Getenv("PORT")
	if port == "" {
		port = "13200"
	}

	port = fmt.Sprintf(":%s", port)
	app.Listen(port)
}
