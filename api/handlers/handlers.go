package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ipinfo/go/v2/ipinfo"
	"golang.org/x/crypto/ssh"

	"github.com/SaidovZohid/swiftsend.it/config"
	"github.com/SaidovZohid/swiftsend.it/pkg/logger"
	"github.com/SaidovZohid/swiftsend.it/sshserver"
	"github.com/SaidovZohid/swiftsend.it/storage"
	"github.com/gofiber/fiber/v2"
)

var UserVerifiedHeader = struct {
	LinkOne   string
	LinkTwo   string
	LinkThree string
}{}

var UserNotVerifiedHeader = struct {
	LinkOne   string
	LinkTwo   string
	LinkThree string
}{}

type handlerV1 struct {
	cfg  *config.Config
	log  logger.Logger
	strg storage.StorageI
	// inMemory storage.InMemoryStorageI
	pipes map[string]sshserver.Tunnel
}

type HandlerV1Options struct {
	Cfg  *config.Config
	Log  logger.Logger
	Strg storage.StorageI
	// InMemory storage.InMemoryStorageI
	Pipes map[string]sshserver.Tunnel
}

func New(options *HandlerV1Options) *handlerV1 {
	UserNotVerifiedHeader = struct {
		LinkOne   string
		LinkTwo   string
		LinkThree string
	}{
		LinkOne:   options.Cfg.BaseURL,
		LinkTwo:   options.Cfg.BaseURL + "/signup",
		LinkThree: options.Cfg.BaseURL + "/login",
	}
	UserVerifiedHeader = struct {
		LinkOne   string
		LinkTwo   string
		LinkThree string
	}{
		LinkOne:   options.Cfg.BaseURL,
		LinkTwo:   options.Cfg.BaseURL + "/s/settings/account",
		LinkThree: options.Cfg.BaseURL + "/logout",
	}
	return &handlerV1{
		cfg:  options.Cfg,
		log:  options.Log,
		strg: options.Strg,
		// inMemory: options.InMemory,
		pipes: options.Pipes,
	}
}

func (h *handlerV1) getGithubAccessToken(code, clientID, secretKey string) (string, error) {

	// Set us the request body as JSON
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": secretKey,
		"code":          code,
	}
	requestJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		return "", err
	}

	// POST request to set URL
	req, reqerr := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)
	if reqerr != nil {
		return "", reqerr
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get the response
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return "", reqerr
	}

	// Response body converted to stringified JSON
	respbody, _ := io.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghresp githubAccessTokenResponse
	err = json.Unmarshal(respbody, &ghresp)
	if resperr != nil {
		return "", err
	}

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return ghresp.AccessToken, nil
}

type User struct {
	Username string  `json:"login"`
	Name     string  `json:"name"`
	Email    *string `json:"email"`
}

func (h *handlerV1) getGithubData(accessToken string) (*User, error) {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		return nil, reqerr
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		return nil, reqerr
	}

	// Read the response as a byte slice
	respbody, _ := io.ReadAll(resp.Body)

	var data User
	err := json.Unmarshal(respbody, &data)
	if err != nil {
		return nil, reqerr
	}

	// Convert byte slice to string and return
	return &data, nil
}

func (h *handlerV1) getUserInfoFromGoogle(token string) (*User, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return nil, err
	}

	userdata, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	userinfo := make(map[string]interface{}, 0)
	err = json.Unmarshal(userdata, &userinfo)
	if err != nil {
		return nil, err
	}

	email := userinfo["email"].(string)
	name, ok := userinfo["name"].(string)
	if !ok {
		name = userinfo["given_name"].(string)
	}

	var data User
	data.Email = &email
	data.Username = name
	data.Name = fmt.Sprintf("%v %v", userinfo["given_name"].(string), userinfo["family_name"].(string))

	return &data, nil
}

func ExtractPublicKeyAndFingerprint(authorizedKey string) (string, string, error) {
	// Parse the SSH authorized key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey))
	if err != nil {
		return "", "", err
	}

	// Extract the public key
	pubKeyBytes := pubKey.Marshal()

	// Calculate the fingerprint
	fingerprint := ssh.FingerprintSHA256(pubKey)

	return string(pubKeyBytes), fingerprint[7:], nil
}

// func ExtractTheKeyOfPublic(key string) string {
// 	return strings.Split(key, " ")[1]
// }

func (h *handlerV1) SetCookie(c *fiber.Ctx, name, val string, expires time.Time) {
	if h.cfg.BaseURL == "http://localhost:3000" {
		c.Cookie(&fiber.Cookie{
			Name:     h.cfg.AuthCookieName,
			Value:    val,
			Path:     "/",
			Expires:  expires,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		})
	} else {
		c.Cookie(&fiber.Cookie{
			Name:     h.cfg.AuthCookieName,
			Domain:   "." + h.cfg.BaseURL[8:],
			Value:    val,
			Path:     "/",
			Expires:  expires,
			HTTPOnly: true,
			Secure:   true,
			SameSite: fiber.CookieSameSiteLaxMode,
		})
	}
}

type LocationInfo struct {
	IP       string `json:"ip"`
	Timezone string `json:"timezone"`
}

func GetLocation(ipaddress string, cfg *config.Config) (*LocationInfo, error) {
	// params: httpClient, cache, token. `http.DefaultClient` and no cache will be used in case of `nil`.
	client := ipinfo.NewClient(nil, nil, cfg.LocationInfoKey)

	info, err := client.GetIPInfo(net.ParseIP(ipaddress))
	if err != nil {
		log.Fatal(err)
	}

	locationInfo := LocationInfo{
		Timezone: info.Timezone,
		IP:       info.IP.String(),
	}

	return &locationInfo, nil
}

func removeHTTPProtocol(url string) string {
	// Check if the URL starts with "https://"
	if strings.HasPrefix(url, "https://") {
		return strings.TrimPrefix(url, "https://")
	}

	// Check if the URL starts with "http://"
	if strings.HasPrefix(url, "http://") {
		return strings.TrimPrefix(url, "http://")
	}

	// If no protocol is found, return the original URL
	return url
}
