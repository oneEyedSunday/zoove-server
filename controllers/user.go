package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"zoove/db"
	"zoove/platforms"
	"zoove/types"
	"zoove/util"

	"github.com/gofiber/fiber"
	"github.com/google/uuid"
)

type User struct {
	DB *db.PrismaClient
}

func NewUserHandler(db *db.PrismaClient) *User {
	return &User{DB: db}
}

func (user *User) AuthorizeUser(ctx *fiber.Ctx) {
	rnid, _ := uuid.NewRandom()
	randomid := rnid.String()
	platform := strings.ToLower(ctx.Params("platform"))
	authcode := ctx.Query("code")
	if platform == util.HostDeezer {
		token, err := platforms.HostDeezerUserAuth(authcode)
		if err != nil {
			log.Println("Error authenticating using on deezer")
			log.Println(err)
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error Authing user", "error": err.Error(), "status": http.StatusInternalServerError})
			return
		}
		profile, err := platforms.HostDeezerFetchUserProfile(token)
		if err != nil {
			log.Println("Error fetching user profile")
			util.InternalServerError(ctx, err)
			return
		}
		ctx.Locals("token", token)
		platformid := strconv.FormatInt(profile.ID, 10)

		claims := &types.Token{
			Platform:      util.HostDeezer,
			PlatformID:    platformid,
			PlatformToken: token,
			UUID:          randomid,
		}

		existing, err := user.DB.User.FindOne(db.User.Email.Equals(profile.Email)).Exec(context.Background())
		if err != nil {
			log.Println("Couldnot get existing user from DB")
			log.Printf("Error is: %#v", err)
			if err == db.ErrNotFound {
				signedJWT, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
				out := map[string]interface{}{
					"token": signedJWT,
				}

				if err != nil {
					panic(err)
				}
				log.Println("User does not exist. should create now")
				_, err = user.DB.User.CreateOne(
					db.User.UpdatedAt.Set(time.Now()),
					db.User.FullName.Set(fmt.Sprintf("%s %s", profile.FirstName, profile.LastName)),
					db.User.FirstName.Set(profile.FirstName),
					db.User.LastName.Set(profile.LastName),
					db.User.Country.Set(profile.Country),
					db.User.Lang.Set(profile.Language),
					db.User.UUID.Set(randomid),
					db.User.Email.Set(profile.Email),
				).Exec(context.Background())
				if err != nil {
					log.Println("Error saving new user")
					log.Println(err)
					util.BadRequest(ctx, err)
					return
				}

				util.RequestCreated(ctx, out)
				return
			}
		}
		claims.UUID = existing.UUID
		signedJWT, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
		out := map[string]interface{}{
			"token": signedJWT,
		}

		if err != nil {
			panic(err)
		}
		util.RequestOk(ctx, out)
		return
	} else if platform == util.HostSpotify {
		// spotifyRedirect := os.Getenv("SPOTIFY_REDIRECT_URI")
		spotify, err := platforms.HostSpotifyUserAuth(authcode)
		if err != nil {
			log.Println("Error getting user")
			log.Println(err)
			util.InternalServerError(ctx, err)
			return
		}
		claims := &types.Token{
			Platform:      util.HostSpotify,
			PlatformID:    spotify.ID,
			PlatformToken: "",
			UUID:          randomid,
		}

		// ctx.Locals("token")
		existing, err := user.DB.User.FindOne(db.User.Email.Equals(spotify.Email)).Exec(context.Background())

		if err != nil {
			log.Println("Error finding from the record")
			if err == db.ErrNotFound {
				signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
				out := map[string]string{
					"token": signedJwt,
				}

				log.Println("User does not exist. create new")
				_, err = user.DB.User.CreateOne(
					db.User.UpdatedAt.Set(time.Now()),
					db.User.FullName.Set(""),
					db.User.FirstName.Set(""),
					db.User.LastName.Set(""),
					db.User.Country.Set(spotify.Country),
					db.User.Lang.Set("en"),
					db.User.UUID.Set(randomid),
					db.User.Email.Set(spotify.Email),
				).Exec(context.Background())

				if err != nil {
					log.Println("Error creating new user")
					log.Println(err)
					util.InternalServerError(ctx, err)
					return
				}
				log.Println("Done with this now....")
				util.RequestCreated(ctx, out)
				return
			}
		}

		claims.UUID = existing.UUID
		signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
		out := map[string]string{
			"token": signedJwt,
		}

		util.RequestOk(ctx, out)
		return
	}
}

func (user *User) GetUserProfile(ctx *fiber.Ctx) {
	uuid := ctx.Locals("uuid").(string)
	existing, err := user.DB.User.FindOne(db.User.UUID.Equals(uuid)).Exec(context.Background())
	if err != nil {
		log.Println("Error getting profile from DB")
		log.Println(err)
		if err == db.ErrNotFound {
			util.NotFound(ctx)
			return
		}
	}
	util.RequestOk(ctx, existing)
	return
}
