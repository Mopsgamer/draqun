package routes

import (
	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

func RegisterPagesRoutes(app *fiber.App, db *goqu.Database) {
	app.Get("/",
		func(ctx fiber.Ctx) error {
			return ctx.Render("homepage", controller.MapPage(ctx, &fiber.Map{"Title": "Homepage", "IsHomePage": true}), "partials/main")
		},
	)
	app.Get("/terms",
		func(ctx fiber.Ctx) error {
			return ctx.Render("terms", controller.MapPage(ctx, &fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
	)
	app.Get("/privacy",
		func(ctx fiber.Ctx) error {
			return ctx.Render("privacy", controller.MapPage(ctx, &fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
	)
	app.Get("/acknowledgements",
		func(ctx fiber.Ctx) error {
			return ctx.Render("acknowledgements", controller.MapPage(ctx, &fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Get("/settings",
		func(ctx fiber.Ctx) error {
			user := fiber.Locals[database.User](ctx, perms.LocalAuth)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return ctx.Render("settings", controller.MapPage(ctx, &fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	app.Get("/chat",
		func(ctx fiber.Ctx) error {
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Home", "IsChatPage": true}))
		},
	)
	app.Get("/chat/groups/:group_id",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[database.Member](ctx, perms.LocalMember)
			if member.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": group.Moniker, "IsChatPage": true}))
		},
		perms.GroupById(db, "group_id"),
	)
	app.Put("/groups/:group_id/join",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[database.Member](ctx, perms.LocalMember)
			if member.IsEmpty() {
				// never been a member
				member = database.NewMember(db, group.Id, member.UserId, "")
				if !member.Insert() {
					return htmx.ErrDatabase
				}
			} else if member.IsDeleted {
				// been a member, now isn't
				member.IsDeleted = false
				if !member.Update() {
					return htmx.ErrDatabase
				}
			} else {
				// already a member
				return htmx.ErrUseless
			}

			if !member.JoinActed() {
				return htmx.ErrDatabase
			}

			if htmx.IsHtmx(ctx) {
				// controller_ws.UserSessionMap.Push(
				// 	filter,
				// 	controller.WrapOob("beforeend:#chat", &str),
				// 	controller_ws.SubForMessages,
				// )

				htmx.Redirect(ctx, controller.PathRedirectGroup(group.Id))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		perms.GroupById(db, "group_id"),
	)
	app.Get("/chat/groups/join/:group_name",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[database.Member](ctx, perms.LocalMember)
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
			if group.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			if member.IsEmpty() {
				return ctx.Redirect().To(controller.PathRedirectGroup(group.Id))
			}

			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Join " + group.Moniker, "IsChatPage": true}))
		},
		perms.GroupByName(db, "group_name"),
	)
}
