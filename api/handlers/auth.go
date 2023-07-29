package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/SaidovZohid/swiftsend.it/pkg/utils"
	"github.com/SaidovZohid/swiftsend.it/storage/mongodb"
	"github.com/gofiber/fiber/v2"
	"github.com/mssola/useragent"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	GoogleProvider = "google"
	GithubProvider = "github"
)

// Github sign in or up redirect user to github
func (h *handlerV1) HandleGithubAuth(c *fiber.Ctx) error {
	// Create the dynamic redirect URL for login
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s",
		h.cfg.Github.ClientID,
		h.cfg.Github.RedirectURI,
	)

	return c.Redirect(redirectURL, 301)
}

// Callback for github get code in query by github and authicated access token and get user data and  handle sign in or sign up
func (h *handlerV1) HandleGithubAuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Redirect("/", 301)
	}
	githubAccessToken, err := h.getGithubAccessToken(code, h.cfg.Github.ClientID, h.cfg.Github.SecretKey)
	if err != nil {
		return err
	}
	githubData, err := h.getGithubData(githubAccessToken)
	if err != nil {
		return err
	}

	if githubData == nil {
		return errors.New("github data is empty")
	}

	var id string
	user, err := h.strg.User().FindUserByUsernameForLS(context.Background(), githubData.Username, GithubProvider)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		h.log.Error(err)
		return errors.New("something went unexpected")
	}
	if user == nil {
		userID, err := h.strg.User().RegisterUserFirst(context.Background(), &mongodb.User{
			Username:               githubData.Username,
			Fullname:               githubData.Name,
			Email:                  githubData.Email,
			CreatedAt:              time.Now().Format(time.RFC3339),
			LogInAndSignUpProvider: GithubProvider,
		})
		if err != nil {
			return errors.New("failed to create new user")
		}
		a := userID.(primitive.ObjectID)
		id = a.Hex()
	} else {
		id = user.Id.Hex()
	}

	token, _, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		UserID:   id,
		Duration: time.Hour * 48,
		Username: githubData.Username,
	})
	if err != nil {
		return errors.New("failed to create jwt token, try again")
	}

	// Parse the User-Agent string using the user_agent library
	ua := useragent.New(c.Get("User-Agent"))

	// Get device and OS details
	device := "Desktop"
	if ua.Mobile() {
		device = "Mobile"
	}
	os := ua.OS()
	// Get browser details
	browserName, browserVersion := ua.Browser()

	// Get client IP address using X-Forwarded-For header
	ipAddress := c.Get("X-Forwarded-For")
	locationInfo, err := GetLocation(ipAddress, h.cfg)
	if err != nil {
		log.Println("Failed to get user info: ", err)
	}

	s := mongodb.Session{
		AccessToken: token,
		UserID:      id,
		IpAddress:   ipAddress,
		Device:      fmt.Sprintf("%v, %v, %v-%v", device, os, browserName, browserVersion),
	}
	if locationInfo != nil {
		s.Timezone = locationInfo.Timezone
		s.IpAddress = locationInfo.IP
	}

	sessionID, err := h.strg.Session().CreateSession(context.Background(), &s)
	if err != nil {
		return err
	}

	// Set cookie for 4 days
	h.SetCookie(c, h.cfg.AuthCookieName, sessionID, time.Now().Add(time.Hour*96))

	return c.Redirect(c.BaseURL() + "/s/settings/account")
}

// Google sign in or up redirect user to google
func (h *handlerV1) HandleGoogleAuth(c *fiber.Ctx) error {
	url := h.cfg.Google.Conf.AuthCodeURL("randomstate")

	return c.Redirect(url, 307)
}

// Callback for google get code in query by google and authicated access token and get user data and  handle sign in or sign up
func (h *handlerV1) HandleGoogleCallback(c *fiber.Ctx) error {
	if c.Query("state") != "randomstate" {
		return errors.New("user denied sign in or up")
	}

	code := c.Query("code")

	token, err := h.cfg.Google.Conf.Exchange(context.Background(), code)
	if err != nil {
		return err
	}

	data, err := h.getUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		return err
	}

	var id string
	user, err := h.strg.User().FindUserByEmail(context.Background(), data.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		h.log.Error(err)
		return errors.New("something went unexpected")
	}
	if user == nil {
		userID, err := h.strg.User().RegisterUserFirst(context.Background(), &mongodb.User{
			Username:               data.Username,
			Fullname:               data.Name,
			Email:                  data.Email,
			CreatedAt:              time.Now().Format(time.RFC3339),
			LogInAndSignUpProvider: GoogleProvider,
		})
		if err != nil {
			return errors.New("failed to create new user")
		}
		a := userID.(primitive.ObjectID)
		id = a.Hex()
	} else {
		id = user.Id.Hex()
	}

	accessToken, _, err := utils.CreateToken(h.cfg, &utils.TokenParams{
		UserID:   id,
		Duration: time.Hour * 48,
		Username: data.Username,
	})
	if err != nil {
		return errors.New("failed to create jwt token, try again")
	}

	sessionID, err := h.strg.Session().CreateSession(context.Background(), &mongodb.Session{
		AccessToken: accessToken,
		UserID:      id,
	})
	if err != nil {
		return errors.New("failed to create session, try again")
	}

	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.AuthCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 48),
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	return c.Redirect(c.BaseURL() + "/s/settings/account")
}

// Render sign up page
func (h *handlerV1) HandleSignUpPage(c *fiber.Ctx) error {
	_, id := h.getAuth(c)
	if id != "" {
		return c.Redirect(c.BaseURL())
	}
	return c.Render("signup/index", fiber.Map{
		"links": UserNotVerifiedHeader,
	})
}

// Render sign in page
func (h *handlerV1) HandleLoginPage(c *fiber.Ctx) error {
	_, id := h.getAuth(c)
	if id != "" {
		return c.Redirect(c.BaseURL() + "/s/settings/account")
	}

	return c.Render("login/index", fiber.Map{
		"links": UserNotVerifiedHeader,
	})
}

// Handle log out from website
func (h *handlerV1) HandleLogout(c *fiber.Ctx) error {
	_, id := h.getAuth(c)
	if id == "" {
		return c.Redirect(c.BaseURL() + "/")
	}
	err := h.strg.Session().DeleteSessionByID(context.Background(), id)
	if err != nil {
		h.log.Error(err)
	}

	// It will set the time on 2009 year that is why cookie will automatically deleted
	h.SetCookie(c, h.cfg.AuthCookieName, "", time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))

	return c.Redirect(c.BaseURL()+"/login", 301)
}

// Render landing page of website https://zohiddev.me
func (h *handlerV1) HandleLandingPage(c *fiber.Ctx) error {
	data, _ := h.getAuth(c)

	if data == nil {
		return c.Render("landing/index", nil)
	}
	return c.Render("landing/index", fiber.Map{
		"username": data.Username,
	})
}
