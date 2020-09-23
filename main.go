package main

import (
	"log"
	"net/http"
	"os"
	"zoove/controllers"
	"zoove/db"
	"zoove/middleware"
	"zoove/types"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	jwtware "github.com/gofiber/jwt"
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

	userHandler := controllers.NewUserHandler(client)
	jaeger := controllers.NewJaeger(pool)
	authentication := middleware.NewAuthUserMiddleware(client)

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodOptions, http.MethodDelete},
	}))
	app.Static("/", "./client/build")
	app.Get("/:platform/oauth", userHandler.AuthorizeUser)

	app.Use(jwtware.New(
		jwtware.Config{SigningKey: []byte(os.Getenv("JWT_SECRET")),
			Claims:     &types.Token{},
			ContextKey: "user",
		}))
	app.Use(authentication.AuthenticateUser)
	app.Get("/api/v1.1/me", userHandler.GetUserProfile)
	app.Use(middleware.ExtractInfoMetadata)
	app.Get("/api/v1.1/search", jaeger.JaegerHandler)

	app.Get("/api/v1", func(ctx *fiber.Ctx) {
		ctx.Status(http.StatusOK).Send("Hi")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "13200"
	}
	app.Listen(port)
}
