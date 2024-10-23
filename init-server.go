package main

import (
	"restapp/restapp"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func InitServer() (*fiber.App, error) {
	db, err := InitDB()
	if err != nil {
		return nil, err
	}

	app := fiber.New(fiber.Config{
		Views:             InitVE(),
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// next code groups should be separated into different functions.
	// + should avoid code repeating

	// static
	// idk how to do it automatically,
	// fiber's static methods (file hosting) wont work with js and css
	app.Get("/static/js/htmx.min.js", func(c fiber.Ctx) error {
		return c.SendFile("./web/static/js/htmx.min.js")
	})
	app.Get("/static/js/json-enc.js", func(c fiber.Ctx) error {
		return c.SendFile("./web/static/js/json-enc.js")
	})
	app.Get("/static/css/main.css", func(c fiber.Ctx) error {
		return c.SendFile("./web/static/css/main.css")
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})
	app.Get("/api", func(c fiber.Ctx) error {
		return c.Render("api", fiber.Map{})
	})
	app.Get("/login", func(c fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})
	app.Get("/register", func(c fiber.Ctx) error {
		return c.Render("register", fiber.Map{})
	})

	// post
	app.Post("/register", func(c fiber.Ctx) error {
		rc := restapp.Responder{Ctx: c}
		return rc.UserRegister(db)
	})
	app.Post("/login", func(c fiber.Ctx) error {
		rc := restapp.Responder{Ctx: c}
		return rc.UserLogin(db)
	})

	return app, nil
}
