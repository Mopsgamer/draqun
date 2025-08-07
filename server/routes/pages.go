package routes

import (
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RoutePages(app *fiber.App, db *database.DB) {
	app.Route("/").Get(
		func(ctx fiber.Ctx) error {
			return ctx.Render("homepage", MapPage(ctx, fiber.Map{"Title": "Homepage", "IsHomePage": true}), "partials/main")
		},
	)
	app.Route("/terms").Get(
		func(ctx fiber.Ctx) error {
			return ctx.Render("terms", MapPage(ctx, fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
	)
	app.Route("/privacy").Get(
		func(ctx fiber.Ctx) error {
			return ctx.Render("privacy", MapPage(ctx, fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
	)
	app.Route("/acknowledgements").Get(
		func(ctx fiber.Ctx) error {
			return ctx.Render("acknowledgements", MapPage(ctx, fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Route("/settings").Get(
		func(ctx fiber.Ctx) error {
			user := fiber.Locals[database.User](ctx, perms.LocalAuth)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return ctx.Render("settings", MapPage(ctx, fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	routePagesChat(app, db)
}
