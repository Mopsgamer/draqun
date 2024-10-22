package main

import (
	"log"

	"github.com/gofiber/template/html/v2"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func main() {
	// https://docs.gofiber.io/template/html/
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	api := app.Route("/api")

	// logger
	app.Get("*", logger.New())

	// api
	api.Get(static.New("./api.html"))
	declApi(api)

	// public
	app.Use("/*", static.New("./public")) // wont work

	// main application page
	app.Get("/", func(ctx fiber.Ctx) error {
		ctx.Locals("uri", ctx.OriginalURL())
		return ctx.Render("index", fiber.Map{})
	})

	// listen
	log.Fatal(app.Listen(":3000"))
}

func declApi(api fiber.Register) {

}
