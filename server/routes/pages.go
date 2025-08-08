package routes

import (
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RoutePages(app *fiber.App, db *model.DB) {
	app.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return ctx.Render("homepage", MapPage(ctx, db, fiber.Map{"Title": "Homepage", "IsHomePage": true}), "partials/main")
		},
	)
	app.Get(
		"/terms",
		func(ctx fiber.Ctx) error {
			return ctx.Render("terms", MapPage(ctx, db, fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
	)
	app.Get(
		"/privacy",
		func(ctx fiber.Ctx) error {
			return ctx.Render("privacy", MapPage(ctx, db, fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
	)
	app.Get(
		"/acknowledgements",
		func(ctx fiber.Ctx) error {
			return ctx.Render("acknowledgements", MapPage(ctx, db, fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Get(
		"/settings",
		func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return ctx.Render("settings", MapPage(ctx, db, fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	routePagesChat(app, db)
}
