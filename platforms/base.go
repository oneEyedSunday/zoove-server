package platforms

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"zoove/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
)

// TrackToSearch is a struct that represents a track to search on platforms
type TrackToSearch struct {
	Title   string
	Artiste string
	Pool    *redis.Pool
	// Chan    chan *types.SingleTrack
}

// TrackToSearchChan is a struct similar to TrackToSearch but async by using Chan

// NewTrackToSearch returns a new instance of TrackToSearch
func NewTrackToSearch(title, artiste string, pool *redis.Pool) *TrackToSearch {
	return &TrackToSearch{Artiste: artiste, Title: title, Pool: pool}
}

// AuthorizeUser authorizes the user and returns the user profile
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
		ctx.Locals("token", token)
		util.RequestOk(ctx, profile)
		return
	}
	// url := fmt.Sprintf("%s/oauth/auth.php?app_id=%s&redirect_uri=%s&perms=%s,%s,%s,%s,%s", os.Getenv("DEEZER_AUTH_BASE"), os.Getenv("DEEZER_APP_ID"), os.Getenv("DEEZER_REDIRECT_URI"), util.HostDeezerBasicAccessPermission, util.HostDeezerEmailPermission, util.HostDeezerOfflineAccessPermission, util.HostDeezerManageLibraryAccessPermission, util.HostDeezerListeningHistoryPermission)
}

func CreatePlaylistChan(userID, title, token, platform string, tracks []string, ch chan bool) {

	if platform == util.HostDeezer {
		err := HostDeezerCreatePlaylist(url.QueryEscape(title), userID, token, tracks)
		if err != nil {
			log.Println("Error creating playlist")
			log.Println(err)
			ch <- false
			return
		}
		ch <- true
		return
	} else if platform == util.HostSpotify {
		spotifyTokens, err := HostSpotifyGetAuthorizedAcessToken(token)
		if err != nil {
			log.Println("Error getting correct access token for spotify")
			log.Println(err)
			ch <- false
			return
		}

		err = HostSpotifyCreatePlaylist(userID, title, spotifyTokens.AccessToken, tracks)
		if err != nil {
			log.Println("Error creating spotify playlist")
			log.Println(err)
			ch <- false
			return
		}
		ch <- true
		return
	}
	ch <- false
	return
}

// type PlaylistToSearch struct {
// 	Name
// }
