package util

import (
	"net/http"

	"github.com/gofiber/fiber"
)

const (
	// HostDeezer simply means deezer
	HostDeezer = "deezer"
	// HostSpotify means spotify
	HostSpotify      = "spotify"
	RedisSearchesKey = "searches"
)

// RequestOk sends back a statusOk response to the client.
func RequestOk(ctx *fiber.Ctx, data interface{}) {
	ctx.Status(http.StatusOK).JSON(fiber.Map{"data": data, "message": "Resource found", "error": nil, "status": http.StatusOK})
}

// BadRequest sends back a statusReqBad response to the client
func BadRequest(ctx *fiber.Ctx, err error) {
	ctx.Status(http.StatusBadRequest).Send(fiber.Map{"message": "The request you send is bad", "error": err.Error(), "status": http.StatusBadRequest})
}

// RequestUnAuthorized sends back a statusUnAuthorized to the client
func RequestUnAuthorized(ctx *fiber.Ctx, err error) {
	ctx.Status(http.StatusUnauthorized).Send(fiber.Map{"message": "The request you made is unauthorized", "error": err.Error(), "status": http.StatusUnauthorized})
}

// RequestCreated sends back a statusCreated to the client
func RequestCreated(ctx *fiber.Ctx) {
	ctx.Status(http.StatusCreated).Send(fiber.Map{"message": "The resource has been created", "error": nil, "status": http.StatusCreated})
}

// NotFound sends back a statusNotFound response to the client
func NotFound(ctx *fiber.Ctx) {
	ctx.Status(http.StatusNotFound).Send(fiber.Map{"message": "The resource does not exist", "error": nil, "status": http.StatusNotFound})
}
