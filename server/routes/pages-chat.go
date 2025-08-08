package routes

import (
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func routePagesChat(router fiber.Router, db *model.DB) fiber.Router {
	chat := router.Group("/chat")
	chat.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return ctx.Render("chat", MapPage(ctx, db, fiber.Map{"Title": "Home", "IsChatPage": true}))
		},
	).Name("page.chat")
	chat.Get(
		"/groups/:group_id",
		perms.GroupById(db, "group_id"),
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			if member.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			return ctx.Render("chat", MapPage(ctx, db, fiber.Map{"Title": group.Moniker, "IsChatPage": true}))
		},
	).Name("page.group")
	chat.Get(
		"/groups/join/:group_name",
		perms.GroupByName(db, "group_name"),
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			if group.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			if member.IsAvailable() {
				return ctx.Redirect().To(group.Url(ctx))
			}

			return ctx.Render("chat", MapPage(ctx, db, fiber.Map{"Title": "Join " + group.Moniker, "IsChatPage": true}))
		},
	).Name("page.group.join")
	return chat
}
