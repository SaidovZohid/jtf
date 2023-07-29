package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *handlerV1) HandleDownloadPaage(c *fiber.Ctx) error {
	link := c.Params("link")
	subdomain := c.Params("subdomain")
	if subdomain == "" {
		return c.Render("errors/404", fiber.Map{
			"what": "Subdomain",
			"link": h.cfg.BaseURL,
			"text": "The subdomain you are looking for doesn't exist. But don't worry, you can create it and make it your own by clicking the button below! 🚀✨",
		})
	}

	if subdomain != "unknown" {
		user, err := h.strg.User().FindUserBySubdomain(context.Background(), subdomain)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return c.Render("errors/404", fiber.Map{
				"what": "Subdomain",
				"link": h.cfg.BaseURL,
				"text": "The subdomain you are looking for doesn't exist. But don't worry, you can create it and make it your own by clicking the button below! 🚀✨",
			})
		}
		if user == nil {
			return c.Render("errors/404", fiber.Map{
				"what": "Subdomain and Link",
				"link": h.cfg.BaseURL,
				"text": "The Subdomain and Link you are looking for doesn't exist. But don't worry, you can create subdomain and make it your own by clicking the button below! 🚀✨",
			})
		}
	}

	val, ok := h.pipes[link]
	if !ok {
		// log.Println("Something here")
		return c.Render("errors/404", fiber.Map{
			"what": "File",
			"link": h.cfg.BaseURL,
			"text": "The provided link is either invalid or has already expired. 🚫🔗 Please ensure you have a valid and up-to-date link. ⏳",
		})
	}

	// Create a fixed time zone for GMT+5 (Asia/Tashkent)
	timezone := time.FixedZone("GMT+5", 5*60*60) // 5 hours ahead of UTC

	timeNow := time.Now().In(timezone)

	// Calculate the difference
	diff := timeNow.Sub(val.SentAt)

	// Extract the minutes and seconds from the difference
	minutes := int(diff.Minutes())
	seconds := int(diff.Seconds()) % 60

	linkTime := fmt.Sprintf("%v seconds ago", seconds)
	if minutes != 0 {
		linkTime = fmt.Sprintf("%v minutes %v seconds ago", minutes, seconds)
	}

	name := "Unknown person"
	expireTime := "in 15 minutes"
	var msg *string
	if val.User.Options != nil {
		if val.User.Subdomain != "" && val.User.Options.From != nil {
			name = *val.User.Options.From
		} else if val.User.Subdomain != "" && val.User.Options.From == nil {
			name = val.User.Subdomain
		} else if val.User.Options.From != nil && val.User.Subdomain == "" {
			name = *val.User.Options.From
		}

		if val.User.Options.Save != nil {
			if *val.User.Options.Save == 1 {
				expireTime = fmt.Sprintf("in %v minute", *val.User.Options.Save)
			} else {
				expireTime = fmt.Sprintf("in %v minutes", *val.User.Options.Save)
			}
		}

		if val.User.Options.Message != nil {
			msg = val.User.Options.Message
		}
	}

	if val.User.Subdomain != "" {
		// verified user
		return c.Render("download/index", fiber.Map{
			"is_verified": true,
			"name":        name,
			"link":        h.cfg.BaseURL + "/direct/" + link,
			"link_time":   linkTime,
			"expire_time": expireTime,
			"msg":         msg,
			"base_url":    h.cfg.BaseURL,
		})
	}

	// unverified user
	return c.Render("download/index", fiber.Map{
		"is_verified": false,
		"name":        name,
		"link":        h.cfg.BaseURL + "/direct/" + link,
		"link_time":   linkTime,
		"expire_time": expireTime,
		"msg":         msg,
		"base_url":    h.cfg.BaseURL,
	})
}

type ioWriterReader struct {
	io.Writer
	io.Reader
}

func (h *handlerV1) HandleDirectDownload(c *fiber.Ctx) error {
	link := c.Params("link")
	val, ok := h.pipes[link]
	if !ok {
		return c.Render("errors/404", fiber.Map{
			"what": "File",
			"link": h.cfg.BaseURL,
			"text": "The provided link is either invalid or has already expired. 🚫🔗 Please ensure you have a valid and up-to-date link. ⏳",
		})
	}

	// Create a buffer to hold the zip file contents
	buf := new(bytes.Buffer)

	// Create a zip writer
	zipWriter := zip.NewWriter(buf)

	// Create a new file in the zip archive
	filename := link
	if val.User.Options != nil {
		if val.User.Options.Filename != nil {
			filename = *val.User.Options.Filename
		}
	}
	file, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	// Create a custom io.ReadWriter that reads from and writes to the io.Writer
	ioReadWriter := &ioWriterReader{
		Writer: file,
		Reader: val.File.W.(io.Reader),
	}

	// Copy the content from the io.Writer to the zip file, and notify progress
	_, err = io.Copy(ioReadWriter, ioReadWriter)
	if err != nil {
		return err
	}

	// Close the zip writer
	err = zipWriter.Close()
	if err != nil {
		return err
	}

	// Set the appropriate headers
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, "jtf.zip"))
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Length", fmt.Sprintf("%d", val.File.FileSize))

	// Send the zip file as the response
	err = c.Send(buf.Bytes())
	if err != nil {
		return err
	}
	// Delete the tunnel from the map and send to done channel
	delete(h.pipes, link)
	// close the listening channel
	close(val.DoneChan)

	return nil
}
