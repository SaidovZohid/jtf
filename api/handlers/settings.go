package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SaidovZohid/swiftsend.it/storage/mongodb"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *handlerV1) HandleSettingGetAccount(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)
	user, err := h.strg.User().FindUserByUsername(context.Background(), payload.Username)
	if err != nil {
		h.log.Error(err)
		return err
	}

	if user.Subdomain == nil {
		return c.Render("settings/account", fiber.Map{
			"username": payload.Username,
			"links":    UserVerifiedHeader,
		})
	}

	link := h.cfg.BaseURL + "/domain/" + *user.Subdomain
	if h.cfg.BaseURL != "http://localhost:3000" {
		link = "https://" + *user.Subdomain + "." + h.cfg.BaseURL[8:]
	}

	return c.Render("settings/account", fiber.Map{
		"username": user.Username,
		"link":     link,
		"links":    UserVerifiedHeader,
	})
}

func (h *handlerV1) HandleSettingPostAccount(c *fiber.Ctx) error {
	payload := struct {
		Subdomain string `json:"subdomain"`
		SSHKey    string `json:"sshKey"`
	}{}

	data, _ := h.getAuth(c)

	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	// TODO: check subdomain is valid and is not in restricted subdomains.
	payload.Subdomain = strings.ToLower(payload.Subdomain)

	if strings.Contains(payload.Subdomain, ".") {
		return c.Render("settings/account", fiber.Map{
			"username":  data.Username,
			"error":     "Apologies for the inconvenience! 😊 The subdomain you entered is invalid!",
			"subdomain": payload.Subdomain,
			"key":       payload.SSHKey,
			"links":     UserVerifiedHeader,
		})
	}

	payload.Subdomain = removeHTTPProtocol(payload.Subdomain)

	for _, v := range engshlishNotAllowedSubdomains {
		if v == payload.Subdomain {
			return c.Render("settings/account", fiber.Map{
				"username":  data.Username,
				"error":     "Apologies for the inconvenience! 😊 The subdomain you entered is either invalid or restricted by us. Please double-check the subdomain you provided and try again. If you have any questions or need assistance, feel free to reach out to me(support@zohiddev.me). I'am here to help! 🌟",
				"subdomain": payload.Subdomain,
				"key":       payload.SSHKey,
				"links":     UserVerifiedHeader,
			})
		}
	}

	_, err := h.strg.User().FindUserBySubdomain(context.Background(), payload.Subdomain)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		h.log.Error(err)
		return c.Render("settings/account", fiber.Map{
			"username":  data.Username,
			"error":     "Oops! 😊 The subdomain you entered is already in use! Please choose a different subdomain to continue.",
			"subdomain": payload.Subdomain,
			"key":       payload.SSHKey,
			"links":     UserVerifiedHeader,
		})
	}
	if err == nil {
		return c.Render("settings/account", fiber.Map{
			"username":  data.Username,
			"error":     "Oops! 😊 The subdomain you entered is already in use! Please choose a different subdomain to continue.",
			"subdomain": payload.Subdomain,
			"key":       payload.SSHKey,
			"links":     UserVerifiedHeader,
		})
	}

	_, fingerPrint, err := ExtractPublicKeyAndFingerprint(payload.SSHKey)
	if err != nil {
		h.log.Error(err)
		return c.Render("settings/account", fiber.Map{
			"username": data.Username,
			"error":    "Are you kidding me? 😔 You are not adding valid ssh key!",
			"links":    UserVerifiedHeader,
		})
	}

	err = h.strg.User().SetSubdomainAndSShKey(context.Background(), data.UserID, payload.Subdomain, &mongodb.Keys{
		ID:        primitive.NewObjectID(),
		Name:      fmt.Sprintf(data.Username + "'s key"),
		SSHHash:   fingerPrint,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		h.log.Error(err)
		return err
	}

	return c.Redirect(c.BaseURL() + "/s/settings/account")
}

func (h *handlerV1) HandleSettingGetKeys(c *fiber.Ctx) error {
	data, _ := h.getAuth(c)

	user, err := h.strg.User().FindUserByUsername(context.Background(), data.Username)
	if err != nil {
		h.log.Error(err)
		return err
	}

	keys := make([]struct {
		ID      string
		Name    string
		SSHHash string
	}, 0)
	if len(user.Keys) > 0 {
		for _, v := range user.Keys {
			keys = append(keys, struct {
				ID      string
				Name    string
				SSHHash string
			}{
				ID:      v.ID.Hex(),
				Name:    v.Name,
				SSHHash: v.SSHHash,
			})
		}
	}

	return c.Render("settings/keys", fiber.Map{
		"username": data.Username,
		"keys":     keys,
		"links":    UserVerifiedHeader,
	})
}

func (h *handlerV1) HandleDeleteKey(c *fiber.Ctx) error {
	id := c.Params("id", "")
	if id == "" {
		return c.Redirect(c.BaseURL() + "/s/settings/keys")
	}

	if err := h.strg.User().DeleteKey(context.Background(), id); err != nil {
		return err
	}

	return c.Redirect(c.BaseURL() + "/s/settings/keys")
}

func (h *handlerV1) HandleSettingAddKeyPage(c *fiber.Ctx) error {
	data, _ := h.getAuth(c)
	return c.Render("settings/add_key", fiber.Map{
		"username": data.Username,
		"links":    UserVerifiedHeader,
	})
}

func (h *handlerV1) HandleSettingAddKey(c *fiber.Ctx) error {
	payload := struct {
		Name   string `json:"name"`
		SSHKey string `json:"sshKey"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		h.log.Error(err)
		return err
	}

	data, _ := h.getAuth(c)
	_, fingerPrint, err := ExtractPublicKeyAndFingerprint(payload.SSHKey)
	if err != nil {
		h.log.Error(err)
		return c.Render("settings/add_key", fiber.Map{
			"username": data.Username,
			"error":    "Are you kidding me? 🤕 You are not adding valid ssh key!",
			"links":    UserVerifiedHeader,
		})
	}

	hasTheSameKey := h.strg.User().HasTheSameKey(context.Background(), data.UserID, fingerPrint)
	if hasTheSameKey {
		return c.Render("settings/add_key", fiber.Map{
			"username": data.Username,
			"error":    "Are you kidding me? 🙄 You are adding a key which was already linked to your account!",
			"links":    UserVerifiedHeader,
		})
	}

	err = h.strg.User().PushNewKey(context.Background(), data.UserID, &mongodb.Keys{
		ID:        primitive.NewObjectID(),
		Name:      payload.Name,
		SSHHash:   fingerPrint,
		CreatedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		h.log.Error(err)
		return err
	}

	return c.Redirect(c.BaseURL() + "/s/settings/keys")
}
