package handlers

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *handlerV1) HandleGetSubdomainInfo(c *fiber.Ctx) error {
	subdomain := c.Params("subdomain")
	if subdomain == "" {
		return c.Render("errors/404", fiber.Map{
			"what": "Subdomain",
			"link": h.cfg.BaseURL,
			"text": "The subdomain you are looking for doesn't exist. But don't worry, you can create it and make it your own by clicking the button below! 🚀✨",
		})
	}

	if subdomain == "unknown" || subdomain == "direct" {
		return c.Redirect(h.cfg.BaseURL, 301)
	}

	info, err := h.strg.User().FindUserBySubdomain(context.Background(), subdomain)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return c.Render("errors/404", fiber.Map{
				"what": "Subdomain",
				"link": h.cfg.BaseURL,
				"text": "The subdomain you are looking for doesn't exist. But don't worry, you can create it and make it your own by clicking the button below! 🚀✨",
			})
		}
		return err
	}
	ln := len(info.Keys)

	data, _ := h.getAuth(c)

	// logged in user
	if data != nil {
		user, err := h.strg.User().FindUserByID(context.Background(), data.UserID)
		if err != nil {
			return err
		}

		if user.Subdomain == nil {
			var empty string
			user.Subdomain = &empty
		}

		return c.Render("subdomain/index", fiber.Map{
			"ln":        ln,
			"owns":      *user.Subdomain == subdomain,
			"link":      subdomain + "." + h.cfg.BaseURL[8:],
			"keys":      info.Keys,
			"link_site": h.cfg.BaseURL,
			"username":  user.Username,
			"links":     UserVerifiedHeader,
		})
	}

	// not logged in user
	return c.Render("subdomain/index", fiber.Map{
		"ln":        ln,
		"owns":      false,
		"link":      subdomain + "." + h.cfg.BaseURL[8:],
		"keys":      info.Keys,
		"link_site": h.cfg.BaseURL,
		"links":     UserNotVerifiedHeader,
	})
}
