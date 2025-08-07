package routes

import (
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func routePagesChat(router fiber.Router, db *model.DB) fiber.Register {
	chat := router.Route("/chat")
	chat.Get(
		func(ctx fiber.Ctx) error {
			return ctx.Render("chat", MapPage(ctx, fiber.Map{"Title": "Home", "IsChatPage": true}))
		},
	)
	chat.Route("/groups/:group_id").Get(
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			if member.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			return ctx.Render("chat", MapPage(ctx, fiber.Map{"Title": group.Moniker, "IsChatPage": true}))
		},
		perms.GroupById(db, "group_id"),
	)
	chat.Route("/groups/join/:group_name").Get(
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			if group.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			if member.IsAvailable() {
				return ctx.Redirect().To(group.Url())
			}

			return ctx.Render("chat", MapPage(ctx, fiber.Map{"Title": "Join " + group.Moniker, "IsChatPage": true}))
		},
		perms.GroupByName(db, "group_name"),
	)
	return chat
}
