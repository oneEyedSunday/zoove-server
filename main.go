package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"zoove/controllers"
	"zoove/db"
	"zoove/middleware"
	"zoove/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

var pool *redis.Pool

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

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
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

	app.Get("/deezer/verify", userHandler.VerifyDeezerSignup)
	app.Get("/:platform/oauth", userHandler.AuthorizeUser)
	app.Post("/api/v1.1/user/join", userHandler.AddNewUser)
	app.Use(middleware.ExtractInfoMetadata)
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
