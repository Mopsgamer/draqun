package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func routePagesChat(router fiber.Router) fiber.Router {
	chat := router.Group("/chat", func(ctx fiber.Ctx) error {
		ctx.Locals("IsChatPage", true)
		return ctx.Next()
	})

	chat.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, fiber.Map{"Title": "Home"}))
		},
	).Name("page.chat")

	chat.Get(
		"/groups/:group_id",
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}

			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			_ = member
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, fiber.Map{"Title": group.Moniker}))
		},
	).Name("page.group")

	chat.Get(
		"/groups/join/:group_name",
		perms.GroupByName("group_name"),
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			if !group.IsAvailable() || group.Mode == model.GroupModePrivate {
				return ctx.Redirect().To("/chat")
			}

			if member.IsAvailable() {
				return ctx.Redirect().To(group.Url(ctx))
			}

			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, fiber.Map{"Title": "Join " + group.Moniker}))
		},
	).Name("page.group.join")

	return chat
}
