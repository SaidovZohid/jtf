package api

import (
	"errors"
	"time"

	h "github.com/SaidovZohid/swiftsend.it/api/handlers"
	"github.com/SaidovZohid/swiftsend.it/config"
	"github.com/SaidovZohid/swiftsend.it/pkg/logger"
	"github.com/SaidovZohid/swiftsend.it/sshserver"
	"github.com/SaidovZohid/swiftsend.it/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/django/v3"

	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type RoutetOptions struct {
	Cfg    *config.Config
	Log    logger.Logger
	Engine *django.Engine
	Strg   storage.StorageI
	Pipes  map[string]sshserver.Tunnel
}

func New(opt *RoutetOptions) *fiber.App {
	engine := django.New("www", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views:        engine,
		WriteTimeout: 10 * time.Minute,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			opt.Log.Error(err)
			return c.Render("errors/500", fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Redirect invalid API requests to the main URL
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		return c.Next()
	})

	app.Use(favicon.New(favicon.Config{
		File: "./www/assets/favicon.ico",
		URL:  "/favicon.ico",
	}))
	app.Static("/assets", "./www/assets")

	handlers := h.New(&h.HandlerV1Options{
		Cfg:   opt.Cfg,
		Log:   opt.Log,
		Strg:  opt.Strg,
		Pipes: opt.Pipes,
	})

	app.Get("/", handlers.HandleLandingPage)
	app.Get("/terms", handlers.HandleTermsPage)
	app.Get("/how-to-use", handlers.HandleHowToUsePage)
	app.Get("/login", handlers.HandleLoginPage)         // login page
	app.Get("/signup", handlers.HandleSignUpPage)       // signup page
	app.Get("/logout", handlers.HandleLogout)           // logout api
	app.Get("/login/github", handlers.HandleGithubAuth) // handler github login
	// app.Get("/login/google", handlers.HandleGoogleAuth)                  // handler google login
	app.Get("/login/github/callback", handlers.HandleGithubAuthCallback) // handler github login callback
	// app.Get("/login/google/callback", handlers.HandleGoogleCallback)     // handler github login callback

	// get subdomain info
	app.Get("/domain/:subdomain", handlers.HandleGetSubdomainInfo)

	// download apis
	app.Get("/download/:subdomain/:link", handlers.HandleDownloadPaage)
	app.Get("/direct/:link", handlers.HandleDirectDownload)

	// delete sent file uri
	app.Get("/delete/:link", handlers.HandleDeleteSentFile)

	// error api for checking error page
	app.Get("/error", func(c *fiber.Ctx) error {
		// Simulate an error
		return app.ErrorHandler(c, errors.New("something went wrong"))
	})

	must := app.Group("/s", handlers.AuthMiddleware)
	must.Get("/settings/account", handlers.HandleSettingGetAccount)
	must.Post("/settings/create/account", handlers.HandleSettingPostAccount)
	must.Get("/settings/keys", handlers.HandleSettingGetKeys)
	must.Post("/settings/keys/d/:id", handlers.HandleDeleteKey)
	must.Get("/settings/keys/add", handlers.HandleSettingAddKeyPage)
	must.Post("/settings/keys/add", handlers.HandleSettingAddKey)

	app.Use(func(c *fiber.Ctx) error {
		return c.Redirect("/", fiber.StatusFound)
	})

	return app
}