package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"zoove/types"

	"github.com/gofiber/fiber"
)

func ExtractInfoMetadata(ctx *fiber.Ctx) {
	rawURL := ctx.Query("track")
	if rawURL == "" {
		log.Println("the req is invalid")
		ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "The request is missing important part", "error": "Missing track query parameter"})
		return
	}

	song, err := url.QueryUnescape(rawURL)
	if err != nil {
		log.Println("Error escaping URL")
		ctx.Next(err)
	}
	parsedURL, err := url.Parse(song)
	if err != nil {
		log.Println("Error parsing URL")
		ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error getting parsing the URL", "error": err.Error()})
		return
	}

	platformHost := parsedURL.Host
	index := strings.Index(song, "?")
	sub := ""
	if index == -1 {
		sub = song
	} else {
		sub = song[:index]
	}
	midd := &types.ExtractedInfo{}
	// for deezer, a song is typically like this:A, https://www.deezer.com/en/track/545820622. but to
	// use the API to get song info, its like this:B, https://api.deezer.com/track/3135556.
	// the below code simply uses the url from A and turn it into B

	if platformHost == "www.deezer.com" {
		deezerID := sub[32:]
		midd.Host = "deezer"
		midd.URL = fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
		midd.ID = deezerID
	} else if platformHost == "open.spotify.com" {
		spotifyID := sub[31:]
		midd.Host = "spotify"
		midd.URL = fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), spotifyID)
		midd.ID = spotifyID
	}

	ctx.Locals("extractedInfo", midd)
	ctx.Next()
}
