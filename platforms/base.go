package platforms

import (
	"log"
	"net/http"
	"strings"
	"zoove/util"

	"github.com/gofiber/fiber"
)

func AuthorizeUser(ctx *fiber.Ctx) {
	platform := strings.ToLower(ctx.Params("platform"))

	if platform == util.HostDeezer {
		authcode := ctx.Query("code")
		token, err := HostDeezerUserAuth(authcode)
		if err != nil {
			log.Println("Error authenticating using on deezer")
			log.Println(err)
			ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error Authing user", "error": err.Error(), "status": http.StatusInternalServerError})
			return
		}
		profile, err := HostDeezerFetchUserProfile(token)
		if err != nil {
			log.Println("Error fetching user profile")
			util.InternalServerError(ctx, err)
			return
		}
		util.RequestOk(ctx, profile)
		return
	}
	// url := fmt.Sprintf("%s/oauth/auth.php?app_id=%s&redirect_uri=%s&perms=%s,%s,%s,%s,%s", os.Getenv("DEEZER_AUTH_BASE"), os.Getenv("DEEZER_APP_ID"), os.Getenv("DEEZER_REDIRECT_URI"), util.HostDeezerBasicAccessPermission, util.HostDeezerEmailPermission, util.HostDeezerOfflineAccessPermission, util.HostDeezerManageLibraryAccessPermission, util.HostDeezerListeningHistoryPermission)
}
