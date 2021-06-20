package middleware

import (
	"context"
	"log"
	"net/http"
	"zoove/db"
	"zoove/types"
	"zoove/util"

	// "github.com/dgrijalva/jwt-go"
	jwt "github.com/form3tech-oss/jwt-go"
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
	log.Println("Extracted search URL from the pasted URL is: ", extracted)
	return ctx.Next()
}

type AuthenticateMiddleware struct {
	DB *db.PrismaClient
}

func (auth *AuthenticateMiddleware) AuthenticateUser(ctx *fiber.Ctx) error {
	ten := ctx.Locals("token").(*jwt.Token)
	claims := ten.Claims.(*types.Token)
	ccx := context.TODO()
	user, err := auth.DB.User.FindFirst(db.User.UUID.Equals(claims.UUID)).Exec(ccx)
	if err != nil {
		if err == db.ErrNotFound {
			log.Println("User with that UUID doesnt exist")
			if claims.Role != string(db.RoleUSER) {
				log.Printf("It seems this is an API key and the role is %s\n", claims.Role)
				ctx.Locals("key_role", claims.Role)
				return ctx.Next()
			}
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "User not found", "error": err})
		}
	}

	ctx.Locals("uuid", user.UUID)
	ctx.Locals("key_role", user.Role)
	return ctx.Next()
}

func NewAuthUserMiddleware(db *db.PrismaClient) *AuthenticateMiddleware {
	return &AuthenticateMiddleware{DB: db}
}
