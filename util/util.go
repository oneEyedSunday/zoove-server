package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"zoove/errors"
	"zoove/types"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

const (
	// HostDeezer simply means deezer
	HostDeezer = "deezer"
	// HostSpotify means spotify
	HostSpotify                             = "spotify"
	RedisSearchesKey                        = "searches"
	HostDeezerBasicAccessPermission         = "basic_access"
	HostDeezerEmailPermission               = "email"
	HostDeezerOfflineAccessPermission       = "offline_access"
	HostDeezerManageLibraryAccessPermission = "manage_library"
	HostDeezerManageCommunityPermission     = "manage_community"
	HostDeezerDeleteLibraryPermission       = "delete_library"
	HostDeezerListeningHistoryPermission    = "listening_history"
)

// RequestOk sends back a statusOk response to the client.
func RequestOk(ctx *fiber.Ctx, data interface{}) error {
	return ctx.Status(http.StatusOK).JSON(fiber.Map{"data": data, "message": "Resource found", "error": nil, "status": http.StatusOK})
}

// BadRequest sends back a statusReqBad response to the client
func BadRequest(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "The request you send is bad", "error": err.Error(), "status": http.StatusBadRequest, "data": nil})
}

// RequestUnAuthorized sends back a statusUnAuthorized to the client
func RequestUnAuthorized(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"message": "The request you made is unauthorized", "error": err.Error(), "status": http.StatusUnauthorized, "data": nil})
}

// RequestCreated sends back a statusCreated to the client
func RequestCreated(ctx *fiber.Ctx, data interface{}) error {
	return ctx.Status(http.StatusCreated).JSON(fiber.Map{"message": "The resource has been created", "error": nil, "status": http.StatusCreated, "data": data})
}

// NotFound sends back a statusNotFound response to the client
func NotFound(ctx *fiber.Ctx) error {
	return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"message": "The resource does not exist", "error": nil, "status": http.StatusNotFound, "data": nil})
}

// InternalServerError returns an error 500
func InternalServerError(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Internal Server Error", "error": err, "status": http.StatusInternalServerError, "data": nil})
}

// NotImplementedError returns a not implemented error
func NotImplementedError(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusNotImplemented).JSON(fiber.Map{"message": "Not yet implemented", "error": err, "status": http.StatusNotImplemented, "data": nil})
}

// RequestForbidden returns a forbidden error
func RequestForbidden(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusForbidden).JSON(fiber.Map{"message": "This request is forbidden", "error": err, "status": http.StatusForbidden})
}

// SignJwtToken signs the token that is returned for a user
func SignJwtToken(claims *types.Token, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &types.Token{
		PlatformToken: claims.PlatformToken,
		Platform:      claims.Platform,
		UUID:          claims.UUID,
		Role:          claims.Role,
		PlatformID:    claims.PlatformID,
		// StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 3).Unix()},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// func SignAccessToken(claims *types.Token, secret)

// SignJwtTokenExp signs the token that is returned for a user but sets the expiration to 5 mins
func SignJwtTokenExp(claims *types.Token, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &types.Token{
		PlatformToken:  claims.PlatformToken,
		Platform:       claims.Platform,
		UUID:           claims.UUID,
		PlatformID:     claims.PlatformID,
		Role:           claims.Role,
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(time.Minute * 30).Unix()},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseJwtToken parses a jwt and returns the claims
func ParseJwtToken(value, secret string) (*types.Token, error) {
	tk := &types.Token{}
	tok, err := jwt.ParseWithClaims(value, tk, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.BadOrInvalidJwt
		}
		return []byte(secret), nil
	})

	// log.Printf("User token is valid: %v\n", token.Valid)
	// tp, _ := err.(*jwt.ValidationError)
	// log.Printf("%#v\n", tp.Error())
	if err != nil {
		log.Printf("err: %#v\n", err.Error())
		return nil, errors.BadOrInvalidJwt
	}
	if !tok.Valid {
		return nil, errors.BadOrInvalidJwt
	}
	return tk, nil
}

// ExtractInfoMetadata extracts metadata from URL
func ExtractInfoMetadata(rawURL string) (*types.ExtractedInfo, error) {
	// rawURL := ctx.Query("track")
	song, err := url.QueryUnescape(rawURL)

	if err != nil {
		log.Println("Error escaping URL")
		return nil, err
	}
	parsedURL, err := url.Parse(song)
	if err != nil {
		log.Println("Error parsing URL")
		return nil, err
		// return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Error getting parsing the URL", "error": err.Error()})
	}

	platformHost := parsedURL.Host
	index := strings.Index(song, "?")
	sub := ""
	if index == -1 {
		sub = song
	} else {
		sub = song[:index]
	}
	extracted := &types.ExtractedInfo{}
	// for deezer, a song is typically like this:A, https://www.deezer.com/en/track/545820622. but to
	// use the API to get song info, its like this:B, https://api.deezer.com/track/3135556.
	// the below code simply uses the url from A and turn it into B

	if platformHost == "www.deezer.com" {
		// find index of playlist
		playlistIndex := strings.Index(sub, "playlist")
		deezerID := ""
		queryType := "track"
		if playlistIndex != -1 {
			deezerID = sub[playlistIndex+9:]
			queryType = "playlist"
		} else {
			trackIndex := strings.Index(sub, "track")
			deezerID = sub[trackIndex+6:]
		}
		extracted.Host = "deezer"
		extracted.URL = fmt.Sprintf("%s/track/%s", os.Getenv("DEEZER_API_BASE"), deezerID)
		extracted.ID = deezerID
		extracted.Type = queryType
	} else if platformHost == "open.spotify.com" {
		playlistIndex := strings.Index(sub, "playlist")
		spotifyID := ""
		queryType := "track"
		if playlistIndex != -1 {
			spotifyID = sub[34:]
			queryType = "playlist"
		} else {
			spotifyID = sub[31:]
		}

		extracted.Host = "spotify"
		extracted.URL = fmt.Sprintf("%s/v1/tracks/%s", os.Getenv("SPOTIFY_API_BASE"), spotifyID)
		extracted.ID = spotifyID
		extracted.Type = queryType
	} else {
		log.Println("Oops! doesnt seem to be a valid playlist or track URL")
		extracted = &types.ExtractedInfo{}
	}
	return extracted, nil
}

// EncryptRefreshToken encrypts a refreshToken for a user
func EncryptRefreshToken(refreshToken string) {}

// TODO: implement refresh token encryption

// MakeRequest makes the http request and marshalls the output inside src
func MakeRequest(url string, src interface{}) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	log.Printf("The URL we're calling is: %#v\n", url)
	if err != nil {
		log.Println("Error GETin URL")
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("Error making HTTP req")
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if strings.Contains(string(body), `{"error`) {
		return errors.NotFound
	}

	if err != nil {
		log.Println("Error reading response into memory")
		return err
	}
	log.Printf("Body is: %s", string(body))
	if res.StatusCode == http.StatusUnauthorized {
		return errors.UnAuthorized
	}

	if res.StatusCode == http.StatusInternalServerError {
		return err
	}

	if string(body) == "true" {
		src = true
		return nil
	}

	err = json.Unmarshal(body, src)
	if err != nil {
		log.Println("Error unserializing response into json")
		return err
	}

	return nil
}

func FetchSpotifyAccessToken() (string, error) {
	eencodedValues := url.Values{}
	eencodedValues.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewReader([]byte(eencodedValues.Encode())))
	if err != nil {
		log.Println("Error making POST request")
		log.Println(err)
		return "", nil
	}
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", spotifyClientID, spotifyClientSecret)))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println("Error making POST making request")
		log.Println(err)
		return "", err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusUnauthorized {
		log.Println("Not enough permission to make the request")
		return "", err
	}
	type AuthTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	reqOutput := &AuthTokenResponse{}
	err = json.NewDecoder(bytes.NewReader(data)).Decode(reqOutput)
	if err != nil {
		log.Println("Error decoding reader")
		log.Println(err)
		return "", err
	}
	return reqOutput.AccessToken, nil
}

func RefreshAuthToken(refreshToken string) (string, error) {
	encodedValues := url.Values{}
	encodedValues.Set("grant_type", "refresh_token")
	encodedValues.Set("refresh_token", refreshToken)

	req, err := http.NewRequest(http.MethodPost, "https://accounts.spotify.com/api/token", bytes.NewReader([]byte(encodedValues.Encode())))
	if err != nil {
		log.Println("Error making POST request")
		return "", err
	}

	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", spotifyClientID, spotifyClientSecret)))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Println("Error making POST making request")
		log.Println(err)
		return "", err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode == http.StatusUnauthorized {
		log.Println("Not enough permission to make the request")
		return "", err
	}
	type AuthTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	reqOutput := &AuthTokenResponse{}
	err = json.NewDecoder(bytes.NewReader(data)).Decode(reqOutput)
	if err != nil {
		log.Println("Error decoding reader")
		log.Println(err)
		return "", err
	}

	log.Println("Scopes authorized for user is", reqOutput.Scope)
	return reqOutput.AccessToken, nil
}

func NewEncryptionKey() *[32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return &key
}

func Encrypt(plaintext []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, goerrors.New("malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}
