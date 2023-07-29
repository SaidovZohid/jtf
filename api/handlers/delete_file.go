package handlers

import "github.com/gofiber/fiber/v2"

func (h *handlerV1) HandleDeleteSentFile(c *fiber.Ctx) error {
	link := c.Params("link")
	val, ok := h.pipes[link]
	if !ok {
		// log.Println("Something here")
		return c.Render("errors/404", fiber.Map{
			"what": "File",
			"link": h.cfg.BaseURL,
			"text": "The provided link is either invalid or has already expired. 🚫🔗 Please ensure you have a valid and up-to-date link. ⏳",
		})
	}

	// Delete the tunnel from the map and send to done channel
	delete(h.pipes, link)
	// close the listening channel
	close(val.DeleteChan)

	return c.SendString("File link deleted successfully!")
}
