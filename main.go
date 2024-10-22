package main

import (
	"log"

	"github.com/gofiber/template/html/v2"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	// template engine
	// https://docs.gofiber.io/template/html/
	engine := html.New("./views", ".html")
	engine.Debug(true)
	engine.Reload(true)
	declFuncs(engine)

	// app
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	// logger
	app.Get("*", logger.New())

	// api
	api := app.Route("/api")
	api.Get(static.New("./api.html"))

	// public
	// app.Use("/*", static.New("./public")) // wont work
	app.Get("/public/main.css", static.New("./public/main.css"))
	app.Get("/public/js/htmx.min.js", static.New("./public/js/htmx.min.js"))

	// main application page
	app.Get("/", func(ctx fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"uri":     ctx.Context().URI(),
			"content": "<b>test</b>",
		})
	})

	// listen
	log.Fatal(app.Listen(":3000"))
}
