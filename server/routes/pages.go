package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RoutePages(app *fiber.App, db *model.DB) {
	app.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "homepage", MapPage(ctx, db, fiber.Map{"Title": "Homepage", "IsHomePage": true}), "partials/main")
		},
	)
	app.Get(
		"/terms",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "terms", MapPage(ctx, db, fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
	)
	app.Get(
		"/privacy",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "privacy", MapPage(ctx, db, fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
	)
	app.Get(
		"/acknowledgements",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "acknowledgements", MapPage(ctx, db, fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Get(
		"/settings",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx, db)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return htmx.TryRenderPage(ctx, "settings", MapPage(ctx, db, fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	routePagesChat(app, db)
}
