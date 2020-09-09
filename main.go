package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber"
	"github.com/joho/godotenv"
)

type SpotifyClientAuth struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func loadEnv() {
	envr := os.Getenv("ENV")
	log.Print(os.Getenv("ENV"))
	err := godotenv.Load(".env." + envr)
	if err != nil {
		log.Println("Error reading the env file")
		log.Println(err)
		panic(err)
	}
}

func init() {
	loadEnv()
}

func main() {
	app := fiber.New()

	app.Get("/api/v1", func(ctx *fiber.Ctx) {
		ctx.Status(http.StatusOK).Send("Hi")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "13200"
	}
	app.Listen(port)
}

func GetSpotifyAuthToken() {
	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", nil)
	if err != nil {
		log.Fatalf("Error with spotify auth")
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifySecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	reqBody := url.Values{}
	client := &http.Client{}

	bearer := base64.StdEncoding.EncodeToString([]byte(spotifyClientID + ":" + spotifySecret))
	req.Header.Set("Authorization", "Basic "+bearer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqBody.Set("grant_type", "client_credentials")

	doRequest, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(doRequest.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer doRequest.Body.Close()

	var out interface{}
	err = json.Unmarshal(body, out)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Spotify token is %#v\n", out)
}
