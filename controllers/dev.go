package controllers

import (
	"log"
	"os"
	"zoove/db"
	"zoove/types"
	"zoove/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/soveran/redisurl"
)

type Developer struct {
	Name       string   `json:"string"`
	Email      string   `json:"email"`
	Permission []string `json:"permissions"`
}

type DeveloperHandler struct {
	DB    *db.PrismaClient
	Redis *redis.Pool
}

func NewDeveloperHandler(db *db.PrismaClient, pool *redis.Pool) *DeveloperHandler {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			log.Println(os.Getenv("REDIS_URL"))
			return redisurl.Connect()
		},
	}
	return &DeveloperHandler{DB: db, Redis: pool}
}

func (dev *DeveloperHandler) CreateAccessToken(ctx *fiber.Ctx) error {
	u, _ := uuid.NewRandom()
	uniqueID := u.String()
	tokenClaims := &types.Token{
		Platform:      "GLOBAL",
		PlatformID:    "",
		Role:          "DEVELOPER",
		UUID:          uniqueID,
		PlatformToken: "",
	}
	token, err := util.SignJwtTokenExp(tokenClaims, os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Println("Error generating access token")
		log.Println(err)
		return util.InternalServerError(ctx, err)
	}
	return util.RequestCreated(ctx, token)
}

// func (dev *DeveloperHandler) AdminLogin(ctx *fiber.Ctx) error {
// 	username :=
// }
