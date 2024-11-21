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

	app.Post("/account/create", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserSignUp()
	})
	app.Post("/account/login", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogin()
	})
	app.Post("/groups/create", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		// TODO: endpoint - create group
		return r.GroupCreate()
	})
	app.Post("/groups/:group_id/leave", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		// TODO: endpoint - send message
		return r.MessageCreate()
	})

	// put
	app.Put("/account/change/name", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeName()
	})
	app.Put("/account/change/email", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeEmail()
	})
	app.Put("/account/change/phone", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePhone()
	})
	app.Put("/account/change/password", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePassword()
	})
	app.Put("/account/logout", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogout()
	})
	app.Put("/groups/:group_id/change", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		// TODO: endpoint - change group
		return r.GroupLeave()
	})

	// delete
	app.Delete("/groups/:group_id/leave", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		// TODO: endpoint - leave group
		return r.GroupLeave()
	})
	app.Delete("/groups/:group_id", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		// TODO: endpoint - delete group
		return r.GroupDelete()
	})
	app.Delete("/account/delete", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserDelete()
	})

	// websoket
	// https://docs.gofiber.io/contrib/next/websocket/
	// TODO: ws - update messages
	// TODO: ws - update members

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
