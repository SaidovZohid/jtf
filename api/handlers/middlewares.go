package handlers

import (
	"context"

	"github.com/SaidovZohid/swiftsend.it/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

func (h *handlerV1) AuthMiddleware(c *fiber.Ctx) error {
	payload, _ := h.getAuth(c)
	if payload == nil {
		return c.Redirect(c.BaseURL() + "/login")
	}

	return c.Next()
}

func (h *handlerV1) getAuth(c *fiber.Ctx) (*utils.Payload, string) {
	cookie := c.Cookies(h.cfg.AuthCookieName)
	if cookie == "" {
		return nil, ""
	}

	res, err := h.strg.Session().GetSessionByID(context.Background(), cookie)
	if err != nil {
		return nil, ""
	}

	payload, err := utils.VerifyToken(h.cfg, res.AccessToken)
	if err != nil {
		return nil, ""
	}

	return payload, cookie
}
