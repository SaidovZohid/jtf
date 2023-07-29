package handlers

import "github.com/gofiber/fiber/v2"

func (h *handlerV1) HandleTermsPage(c *fiber.Ctx) error {
	data, _ := h.getAuth(c)
	if data == nil {
		return c.Render("terms/index", fiber.Map{
			"links": UserNotVerifiedHeader,
		})
	}

	return c.Render("terms/index", fiber.Map{
		"username": data.Username,
		"links":    UserVerifiedHeader,
	})
}

func (h *handlerV1) HandleHowToUsePage(c *fiber.Ctx) error {
	data, _ := h.getAuth(c)
	if data == nil {
		return c.Render("how_to_use/index", fiber.Map{
			"links":    UserNotVerifiedHeader,
			"base_url": h.cfg.BaseURL,
		})
	}

	return c.Render("how_to_use/index", fiber.Map{
		"username": data.Username,
		"links":    UserVerifiedHeader,
		"base_url": h.cfg.BaseURL,
	})
}
