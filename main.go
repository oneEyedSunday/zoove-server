package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"zoove/controllers"
	"zoove/db"
	"zoove/middleware"
	"zoove/service"
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

func loadListeners() {
	for {
		select {
		case <-register:
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
	devHandler := controllers.NewDeveloperHandler(client, pool)
	jaeger := controllers.NewJaeger(pool)
	authentication := middleware.NewAuthUserMiddleware(client)

	go loadListeners()

	app.Use(cors.New(cors.Config{
		AllowMethods: fmt.Sprintf("%s,%s,%s,%s,%s", http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodOptions, http.MethodDelete),
		AllowOrigins: "*",
	}))

	type Sample struct {
		AccessToken string `query:"access_token"`
	}

	app.Get("/deezer/channel.html", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).SendFile("./channel.html")
	})

	app.Get("/:platform/join", func(c *fiber.Ctx) error {
		platform := c.Params("platform")
		log.Print("User is trying to join or login")
		log.Println(platform)
		if platform == util.HostDeezer {

			DeezerAuthBase := os.Getenv("DEEZER_AUTH_BASE")
			DeezerAppID := os.Getenv("DEEZER_APP_ID")
			DeezerRedirectURI := os.Getenv("DEEZER_REDIRECT_URI")
			scopes := "basic_access,email,offline_access,listening_history,manage_library"
			url := fmt.Sprintf("%s/auth.php?app_id=%s&redirect_uri=%s&perms=%s", DeezerAuthBase, DeezerAppID, DeezerRedirectURI, scopes)
			return c.Redirect(url)
		} else if platform == util.HostSpotify {
			spotifyAuthBase := os.Getenv("SPOTIFY_AUTH_BASE")
			spotifyAppID := os.Getenv("SPOTIFY_APP_ID")
			spotifyRedirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
			scopes := url.QueryEscape("user-read-private user-read-email playlist-modify-public playlist-modify-private user-library-modify user-top-read user-read-recently-played user-read-currently-playing")
			url := fmt.Sprintf("%s/authorize?response_type=code&client_id=%s&scope=%s&redirect_uri=%s", spotifyAuthBase, spotifyAppID, scopes, spotifyRedirectURI)
			return c.Redirect(url)
		}
		return util.NotImplementedError(c, nil)
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
		pool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return redisurl.Connect()
			},
		}

		register <- c
		for {

			log.Println("Should print something here!! here it is..")
			_, msg, err := c.ReadMessage()
			log.Println("Here is the websocket message sent: ", string(msg))
			if string(msg) == "" {
				log.Println("Empty message... probably means trying to (re)connect")
				c.WriteMessage(websocket.CloseServiceRestart, []byte(`{"desc":"restart", "message":"service restart"`))
				c.Close()
				// listener.c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"error", "message":"Its me not you...."`))
				// listener.c.Close()
			}
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("Read Error:", err)
				}
				return
			}

			deserialize := &service.SocketMessage{}
			err = json.Unmarshal(msg, deserialize)
			if err != nil {
				log.Println("Error parsing. Seems client is sending non-json data")
				c.WriteMessage(websocket.TextMessage, []byte(`{"desc":"send JSON unmarshalling errors here"}`))
				c.Close()
			}

			var trackMeta = &types.SingleTrack{}
			var playlistMeta = &types.Playlist{}
			listener := &service.SocketListener{Deserialize: *deserialize,
				C: c.Conn, Client: client, DeezerTracks: deezerTracks,
				PlaylistMeta:  playlistMeta,
				SpotifyTracks: spotifyTracks,
				TrackMeta:     trackMeta,
				Tracks:        tracks,
			}
			if deserialize.Type == "track" {
				listener.GetTrackListener()
			} else if deserialize.Type == "playlist" {
				listener.GetPlaylistListener()
			} else if deserialize.Type == "create_playlist" {
				listener.CreatePlaylistListener()
			} else {
				log.Println("No specific action type: probably use a catch-all here.")
				c.Close()
			}
		}
	}))
	app.Get("/:platform/signup", userHandler.SignupRedirect)
	app.Get("/deezer/verify", userHandler.VerifyDeezerSignup)
	app.Get("/kanye/:platform/oauth", userHandler.AuthorizeUser)
	app.Post("/api/v1.1/user/join", userHandler.AddNewUser)
	// app.Post("/api/v1/developer/create")
	app.Get("/api/v1.1/kanye/gggghhh/create", devHandler.CreateAccessToken)
	app.Use(middleware.ExtractedInfoMiddleware)
	app.Get("/api/v1.1/zoovify/playlist", jaeger.ConvertPlaylist)
	// app.Get("/api/v1.1/admin/login")
	app.Use(jwtware.New(
		jwtware.Config{SigningKey: []byte(os.Getenv("JWT_SECRET")),
			Claims:     &types.Token{},
			ContextKey: "token",
		}))
	app.Use(authentication.AuthenticateUser)
	app.Get("/api/v1.1/search", jaeger.JaegerHandler)
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
