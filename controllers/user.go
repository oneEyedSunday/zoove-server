package controllers

import (
	"context"
	"encoding/json"
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

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/soveran/redisurl"
)

// User represents blueprint of things needed to perform operations for user
type User struct {
	DB    *db.PrismaClient
	Redis *redis.Pool
}

// NewUserHandler returns a new pointer for user we want to perform operations on
func NewUserHandler(db *db.PrismaClient, pool *redis.Pool) *User {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			log.Println(os.Getenv("REDIS_URL"))
			return redisurl.Connect()
		},
	}
	return &User{DB: db, Redis: pool}
}

// VerifyDeezerSignup verifies the access token a user copied is still valid
func (user *User) VerifyDeezerSignup(ctx *fiber.Ctx) error {
	jwtToken := ctx.Query("token")
	log.Printf("Token is... %s", jwtToken)
	prs, err := util.ParseJwtToken(jwtToken, os.Getenv("JWT_SECRET"))
	if err != nil {
		return util.RequestUnAuthorized(ctx, err)
	}
	// check if this user exists
	existing, err := user.DB.User.FindOne(db.User.UUID.Equals(prs.UUID)).Exec(context.Background())
	if err != nil {
		log.Printf("[ERROR]: Could not find user. User does not exist: %#v\n", err)
		return util.NotFound(ctx)
	}
	claims := &types.Token{
		Platform:      prs.Platform,
		PlatformID:    prs.PlatformID,
		PlatformToken: existing.Token,
		UUID:          existing.UUID,
	}

	signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Println(err)
		return util.InternalServerError(ctx, err)
	}

	out := map[string]interface{}{
		"token": signedJwt,
		"user":  existing,
	}
	return util.RequestOk(ctx, out)
}

// AuthorizeUser authorizes a user and returns token to login
func (user *User) AuthorizeUser(ctx *fiber.Ctx) error {
	rnid, _ := uuid.NewRandom()
	randomid := rnid.String()
	platform := strings.ToLower(ctx.Params("platform"))
	authcode := ctx.Query("code")
	if platform == util.HostDeezer {
		token, err := platforms.HostDeezerUserAuth(authcode)
		if err != nil {
			log.Println("Error authenticating using on deezer")
			log.Println(err)
			return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error Authing user", "error": err.Error(), "status": http.StatusInternalServerError})
		}

		profile, err := platforms.HostDeezerFetchUserProfile(token)
		if err != nil {
			log.Println("Error fetching user profile")
			return ctx.Status(http.StatusInternalServerError).JSON(err)
		}

		ctx.Locals("token", token)
		platformid := strconv.FormatInt(int64(profile.ID), 10)

		claims := &types.Token{
			Platform:      util.HostDeezer,
			PlatformID:    platformid,
			PlatformToken: token,
			UUID:          randomid,
		}

		existing, err := user.DB.User.FindOne(db.User.Email.Equals(profile.Email)).Exec(context.Background())
		if err != nil {
			if err == db.ErrNotFound {
				signedJWT, err := util.SignJwtTokenExp(claims, os.Getenv("JWT_SECRET"))

				if err != nil {
					panic(err)
				}
				plan := ""
				if profile.Status == 1 {
					plan = "free"
				} else if profile.Status == 2 {
					plan = "premium"
				}

				log.Println("User does not exist. should create now")
				_, err = user.DB.User.CreateOne(
					db.User.UpdatedAt.Set(time.Now()),
					db.User.FullName.Set(fmt.Sprintf("%s %s", profile.Firstname, profile.Lastname)),
					db.User.FirstName.Set(profile.Firstname),
					db.User.LastName.Set(profile.Lastname),
					db.User.Country.Set(profile.Country),
					db.User.Lang.Set(profile.Lang),
					db.User.UUID.Set(randomid),
					db.User.Email.Set(profile.Email),
					db.User.Username.Set(profile.Name),
					db.User.Platform.Set(util.HostDeezer),
					db.User.Avatar.Set(profile.Picture),
					db.User.Token.Set(token), // T0DO: ENCRYPT THIS..
					db.User.Plan.Set(plan),
				).Exec(context.Background())
				if err != nil {
					log.Println("Error saving new user")
					log.Println(err)
					return util.BadRequest(ctx, err)
				}

				out := map[string]interface{}{
					"token": signedJWT,
					// "user":  newUser,
				}
				return util.RequestCreated(ctx, out)
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
		return util.RequestOk(ctx, out)
		// ctx.Status(http.StatusNoContent)
	} else if platform == util.HostSpotify {
		spotify, refreshToken, err := platforms.HostSpotifyUserAuth(authcode)
		if err != nil {
			log.Println("Error getting user")
			log.Println(err)
			return util.InternalServerError(ctx, err)
		}
		claims := &types.Token{
			Platform:      util.HostSpotify,
			PlatformID:    spotify.ID,
			PlatformToken: "",
			UUID:          randomid,
		}

		existing, err := user.DB.User.FindOne(db.User.Email.Equals(spotify.Email)).Exec(context.Background())

		if err != nil {
			log.Println("Error finding from the record")
			if err == db.ErrNotFound {
				signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
				out := map[string]string{
					"token": signedJwt,
				}

				ppix := ""
				if len(spotify.Images) == 0 {
					ppix = ""
				} else {
					ppix = spotify.Images[0].URL
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
					db.User.Username.Set(spotify.DisplayName),
					db.User.Platform.Set(util.HostSpotify),
					db.User.Avatar.Set(ppix),
					db.User.Token.Set(refreshToken),
					db.User.Plan.Set(spotify.Product),
				).Exec(context.Background())

				if err != nil {
					log.Println("Error creating new user")
					log.Println(err)
					return util.InternalServerError(ctx, err)
				}
				log.Println("Done with this now....")
				return util.RequestCreated(ctx, out)
			}
		}

		claims.UUID = existing.UUID
		signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
		out := map[string]string{
			"token": signedJwt,
		}

		util.RequestOk(ctx, out)
		return ctx.JSON(out)
	}
	return util.NotImplementedError(ctx, nil)
}

// GetUserProfile updates a user profile
func (user *User) GetUserProfile(ctx *fiber.Ctx) error {
	uuid := ctx.Locals("uuid").(string)
	existing, err := user.DB.User.FindOne(db.User.UUID.Equals(uuid)).Exec(context.Background())
	if err != nil {
		log.Println("Error getting profile from DB")
		log.Println(err)
		if err == db.ErrNotFound {
			return util.NotFound(ctx)
		}
	}
	return util.RequestOk(ctx, existing)
}

// UpdateUserProfile updates user profile
func (user *User) UpdateUserProfile(ctx *fiber.Ctx) error {
	updateInfo := &types.UserProfileUpdate{}
	uuid := ctx.Locals("uuid").(string)
	existing, err := user.DB.User.FindOne(db.User.UUID.Equals(uuid)).Exec(context.Background())
	if err != nil {
		if err == db.ErrNotFound {
			return util.NotFound(ctx)
		}
		return util.InternalServerError(ctx, err)
	}
	// NOTE: we're passing existing.Country because I dont want to allow for country update yet. lang too
	err = user.DB.QueryRaw(`INSERT INTO "User"(id, email, firstName, lastName, fullName, country, lang, username, platform, avatar,token,plan) 
	VALUES($1, $2, $3, $4, $5, $5, $6,$7, $8, $9, $10, $11, $12) ON DO UPDATE SET email= EXCLUDED.email, firstName = EXCLUDED.firstName,
	lastName = EXCLUDED.lastName, lang = EXCLUDED.lang, country = EXCLUDED.country, fullName = EXCLUDED.fullName, platform = EXCLUDED.platform,
	avatar = EXCLUDED.avatar, token = EXCLUDED.token, plan = EXCLUDED.plan`,
		existing.ID, updateInfo.Email, updateInfo.FirstName, updateInfo.LastName, existing.Country, existing.Lang, updateInfo.Username,
		existing.Platform, existing.Avatar, existing.Token, existing.Plan).Exec(context.Background(), updateInfo)

	if err != nil {
		log.Println("Error executing raw SQL query on DB")
		log.Println(err)
		return util.InternalServerError(ctx, err)
	}

	return util.RequestOk(ctx, updateInfo)
}

// GetListeningHistory returns the listening history for a user
func (user *User) GetListeningHistory(ctx *fiber.Ctx) error {
	conn := user.Redis.Get()
	defer conn.Close()
	history := []types.SingleTrack{}
	uuid := ctx.Locals("uuid").(string)

	existing, err := user.DB.User.FindOne(db.User.UUID.Equals(uuid)).Exec(context.Background())
	if err != nil {
		log.Println("Error fetching user from DB")
		return util.InternalServerError(ctx, err)
	}

	if existing.Platform == util.HostDeezer {
		if existing.Token == "" {
			// TODO: reauth user
		}
		history, err = platforms.HostDeezerFetchHistory(existing.Token)
		if err != nil {
			log.Println("Error fetching user deezer history")
			log.Println(err)
			return util.InternalServerError(ctx, err)
		}

	} else if existing.Platform == util.HostSpotify {
		history, err = platforms.HostSpotifyListeningHistory(existing.Token)
		if err != nil {
			log.Printf("Error getting spotify listening history: %#v\n", err)
			return util.InternalServerError(ctx, err)
		}
	}

	key := fmt.Sprintf("user-%s", existing.UUID)
	serialize, err := json.Marshal(history)
	if err != nil {
		log.Println("Error caching for user")
		log.Printf("%#v\n", err)
	}

	_, err = redis.String(conn.Do("GET", key))
	if err != nil {
		log.Printf("Error getting from redis..:%s", err)
	}

	_, err = redis.Int64(conn.Do("DEL", key))
	if err != nil {
		log.Println("Error removing from redis.")
		log.Println(err)
	}

	_, err = redis.String(conn.Do("SET", key, string(serialize)))
	return util.RequestOk(ctx, history)
}

// GetArtistePlayHistory returns the playlist history of the artistes a user has played
func (user *User) GetArtistePlayHistory(ctx *fiber.Ctx) error {
	conn := user.Redis.Get()
	defer conn.Close()
	uuid := ctx.Locals("uuid").(string)
	existing, _ := user.DB.User.FindOne(db.User.UUID.Equals(uuid)).Exec(context.Background())

	if existing.Platform == util.HostDeezer {
		if existing.Token == "" {
			// T0DO: implement auth user
		}
	}
	key := fmt.Sprintf("user-%s", existing.UUID)
	hist := &[]types.SingleTrack{}
	cached, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return util.NotFound(ctx)
		}
	}
	err = json.Unmarshal([]byte(cached), hist)
	if err != nil {
		return util.BadRequest(ctx, err)
	}
	history := []string{}
	for _, track := range *hist {
		history = append(history, track.Artistes...)
	}

	return util.RequestOk(ctx, history)
}

// AddNewUser adds a new user. this is because for example, mobile needs to call this to be able to create user on backend
func (user *User) AddNewUser(ctx *fiber.Ctx) error {
	newUser := &types.NewUser{}
	err := ctx.BodyParser(newUser)
	if err != nil {
		log.Println("Error adding new user to record")
		return util.InternalServerError(ctx, err)
	}
	// check if the user exists
	rand, _ := uuid.NewRandom()
	claims := &types.Token{
		Platform:      newUser.Platform,
		PlatformID:    newUser.PlatformID,
		PlatformToken: "",
		UUID:          rand.String(),
	}

	existing, err := user.DB.User.FindOne(db.User.Email.Equals(newUser.Email)).Exec(context.Background())
	if err != nil {
		if err == db.ErrNotFound {
			log.Println("Not found")
			n, err := user.DB.User.CreateOne(
				db.User.UpdatedAt.Set(time.Now()),
				db.User.FullName.Set(fmt.Sprintf("%s %s", newUser.FirstName, newUser.LastName)),
				db.User.FirstName.Set(newUser.FirstName),
				db.User.LastName.Set(newUser.LastName),
				db.User.Country.Set(newUser.Country),
				db.User.Lang.Set(newUser.Lang),
				db.User.UUID.Set(rand.String()),
				db.User.Email.Set(newUser.Email),
				db.User.Username.Set(newUser.Username),
				db.User.Platform.Set(newUser.Platform),
				db.User.Avatar.Set(newUser.Avatar),
				db.User.Token.Set(newUser.Token),
				db.User.Plan.Set(newUser.Plan),
				db.User.CreatedAt.Set(time.Now()),
			).Exec(context.Background())
			if err != nil {
				log.Printf("[ERROR]: Error creating new user")
				return util.InternalServerError(ctx, err)
			}
			signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
			if err != nil {
				log.Println(err)
				return util.InternalServerError(ctx, err)
			}
			n.Token = ""
			res := map[string]interface{}{
				"token": signedJwt,
				"user":  n,
			}
			return util.RequestCreated(ctx, res)
		}
	}

	signedJwt, err := util.SignJwtToken(claims, os.Getenv("JWT_SECRET"))
	if err != nil {
		return util.InternalServerError(ctx, err)
	}
	res := map[string]interface{}{
		"token": signedJwt,
		"user":  existing,
	}
	return util.RequestOk(ctx, res)
}
