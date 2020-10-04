package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"zoove/db"
	"zoove/types"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func ExtractInfoMetadata(ctx *fiber.Ctx) error {
	rawURL := ctx.Query("track")
	song, err := url.QueryUnescape(rawURL)

	if err != nil {
		log.Println("Error escaping URL")
		return ctx.Next()
	}
	parsedURL, err := url.Parse(song)
	if err != nil {
		log.Println("Error parsing URL")
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error getting parsing the URL", "error": err.Error()})
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
		// find index of playlist
		playlistIndex := strings.Index(sub, "playlist")
		deezerID := ""
		if playlistIndex != -1 {
			deezerID = sub[playlistIndex+9:]
		} else {
			deezerID = sub[32:]
		}
		midd.Host = "deezer"
		midd.URL = fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
		midd.ID = deezerID
	} else if platformHost == "open.spotify.com" {
		// log.Println("Its spotify...")
		// log.Printf("Sub: %s", sub)
		playlistIndex := strings.Index(sub, "playlist")
		spotifyID := ""
		if playlistIndex != -1 {
			spotifyID = sub[34:]
		} else {
			spotifyID = sub[34:]
		}

		midd.Host = "spotify"
		midd.URL = fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), spotifyID)
		midd.ID = spotifyID
	}

	ctx.Locals("extractedInfo", midd)
	return ctx.Next()
}

type AuthenticateMiddleware struct {
	DB *db.PrismaClient
}

func (auth *AuthenticateMiddleware) AuthenticateUser(ctx *fiber.Ctx) error {
	ten := ctx.Locals("user").(*jwt.Token)
	claims := ten.Claims.(*types.Token)
	ccx := context.TODO()
	user, err := auth.DB.User.FindOne(db.User.UUID.Equals(claims.UUID)).Exec(ccx)
	if err != nil {
		if err == db.ErrNotFound {
			log.Println("User with that UUID doesnt exist")
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "User not found", "error": err})
		}
	}

	ctx.Locals("uuid", user.UUID)
	return ctx.Next()
}

func NewAuthUserMiddleware(db *db.PrismaClient) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{DB: db}
}
