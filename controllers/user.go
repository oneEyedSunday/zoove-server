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
	platform := strings.ToLower(ctx.Params("platform"))

	if platform == util.HostDeezer {
		authcode := ctx.Query("code")
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
		existing, err := user.DB.User.FindOne(db.User.Email.Equals(profile.Email)).Exec(context.Background())
		if err != nil {
			log.Println("Couldnot get existing user from DB")
			if err == db.ErrNotFound {
				log.Println("User does not exist. should create now")
				rnid, _ := uuid.NewRandom()
				randomid := rnid.String()
				platformid := strconv.FormatInt(profile.ID, 10)

				claims := &types.Token{
					Platform:      profile.Platform,
					PlatformID:    platformid,
					PlatformToken: token,
					UUID:          randomid,
				}
				signedJWT, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
				if err != nil {
					panic(err)
				}
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

				util.RequestCreated(ctx, signedJWT)
				return
			}
		}
		util.RequestOk(ctx, existing)
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
