package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RoutePages(app *fiber.App) {
	app.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "homepage", MapPage(ctx, fiber.Map{"Title": "Homepage"}), "partials/main")
		},
	)
	app.Get(
		"/terms",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "terms", MapPage(ctx, fiber.Map{"Title": "Terms"}), "partials/main")
		},
	)
	app.Get(
		"/privacy",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "privacy", MapPage(ctx, fiber.Map{"Title": "Privacy"}), "partials/main")
		},
	)
	app.Get(
		"/acknowledgements",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "acknowledgements", MapPage(ctx, fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Get(
		"/docs",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "docs", MapPage(ctx, fiber.Map{"Title": "Docs"}), "partials/main")
		},
	)
	app.Get(
		"/login",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if !user.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}
			return htmx.TryRenderPage(ctx, "login", MapPage(ctx, fiber.Map{"Title": "Log In"}), "partials/main")
		},
	)
	app.Get(
		"/signup",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if !user.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}
			return htmx.TryRenderPage(ctx, "signup", MapPage(ctx, fiber.Map{"Title": "Sign Up"}), "partials/main")
		},
	)
	app.Get(
		"/settings",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return htmx.TryRenderPage(ctx, "settings", MapPage(ctx, fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	app.Get(
		"/settings/email",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}
			return htmx.TryRenderPage(ctx, "settings-email", MapPage(ctx, fiber.Map{"Title": "Change Email"}), "partials/main")
		},
	)
	app.Get(
		"/settings/password",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}
			return htmx.TryRenderPage(ctx, "settings-password", MapPage(ctx, fiber.Map{"Title": "Change Password"}), "partials/main")
		},
	)
	app.Get(
		"/settings/phone",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}
			return htmx.TryRenderPage(ctx, "settings-phone", MapPage(ctx, fiber.Map{"Title": "Change Phone"}), "partials/main")
		},
	)
	app.Get(
		"/settings/delete",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}
			return htmx.TryRenderPage(ctx, "settings-delete", MapPage(ctx, fiber.Map{"Title": "Delete Account"}), "partials/main")
		},
	)
	app.Get(
		"/groups/create",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}
			return htmx.TryRenderPage(ctx, "group-create", MapPage(ctx, fiber.Map{"Title": "Create Group"}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings",
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			return htmx.TryRenderPage(ctx, "group-settings", MapPage(ctx, fiber.Map{"Title": "Group Settings - " + group.Moniker}), "partials/main")
		},
	)
	routePagesChat(app)
}
