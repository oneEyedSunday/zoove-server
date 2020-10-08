package middleware

import (
	"context"
	"log"
	"net/http"
	"zoove/db"
	"zoove/types"
	"zoove/util"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func ExtractedInfoMiddleware(ctx *fiber.Ctx) error {
	rawURL := ctx.Query("track")
	extracted, err := util.ExtractInfoMetadata(rawURL)
	if err != nil {
		log.Println("Error extracting metadata info")
		log.Println(err)
		return util.BadRequest(ctx, err)
	}
	ctx.Locals("extractedInfo", extracted)
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
