package controllers

import (
	"log"
	"net/http"
	"strings"
	"zoove/db"
	"zoove/platforms"
	"zoove/util"

	"github.com/gofiber/fiber"
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
		util.RequestOk(ctx, profile)
		return
	}
}

func (user *User) GetUserProfile(ctx *fiber.Ctx) {

}
