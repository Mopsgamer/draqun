package main

import (
	"restapp/restapp"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
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
	app.Get("/static/*", static.New("./web/static"))

	// get
	app.Get("/", func(c fiber.Ctx) error {
		return c.Render("index", fiber.Map{}, "layouts/main")
	})
	app.Get("/api", func(c fiber.Ctx) error {
		return c.Render("api", fiber.Map{}, "layouts/main")
	})
	app.Get("/login", func(c fiber.Ctx) error {
		return c.Render("login", fiber.Map{}, "layouts/main")
	})
	app.Get("/register", func(c fiber.Ctx) error {
		return c.Render("register", fiber.Map{}, "layouts/main")
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
