package internal

import (
	"restapp/internal/environment"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Initialize gofiber application, including DB and view engine.
func NewApp() (*fiber.App, error) {
	environment.WaitForBuild()

	db, err := InitDB()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	app := fiber.New(fiber.Config{
		Views:             NewAppHtmlEngine(),
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// next code groups should be separated into different functions.
	// + should avoid code repeating

	// static
	app.Get("/static/*", static.New("./web/static", static.Config{Browse: true}))
	app.Get("/assets/*", static.New("./web/assets", static.Config{Browse: true}))
	app.Get("/partials", static.New("./web/templates/partials", static.Config{Browse: true}))
	app.Get("/partials/*", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderTemplate()
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"",
			"index",
			fiber.Map{
				"Title": "Restapp - Home page",
			},
			"partials/main",
		)
	})
	app.Get("/settings", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"/",
			"settings",
			fiber.Map{
				"Title": "Restapp - Settings",
			},
			"partials/main",
		)
	})
	app.Get("/chat", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"",
			"chat",
			fiber.Map{
				"Title":      "Restapp - Chat",
				"IsChatPage": true,
			},
		)
	})

	// post
	app.Post("/register", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserRegister()
	})
	app.Post("/login", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogin()
	})

	// put
	app.Put("/change-name", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeName()
	})
	app.Put("/change-email", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeEmail()
	})
	app.Put("/change-phone", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePhone()
	})
	app.Put("/change-password", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePassword()
	})
	app.Put("/logout", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogout()
	})

	// delete
	app.Delete("/account-delete", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserDelete()
	})

	app.Use(func(c fiber.Ctx) error {
		r := Responder{c, *db}
		r.Status(fiber.StatusNotFound)
		return r.RenderPage(
			"",
			"partials/x",
			fiber.Map{
				"Title":         "Restapp - " + strconv.Itoa(fiber.StatusNotFound),
				"StatusCode":    fiber.StatusNotFound,
				"StatusMessage": fiber.ErrNotFound.Message,
			},
			"partials/main",
		)
	})

	return app, nil
}
