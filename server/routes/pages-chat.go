package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func routePagesChat(router fiber.Router) fiber.Router {
	chat := router.Group(
		"/chat",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if user.IsEmpty() {
				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, "/login")
					return ctx.Render("partials/redirecting", fiber.Map{})
				}
				return htmx.TryRenderPage(ctx, "chat-login", MapPage(ctx, fiber.Map{"Title": "Home"}), "partials/layout-chat")
			}
			return ctx.Next()
		},
	)

	chat.Get(
		"/",
		func(ctx fiber.Ctx) error {
			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, fiber.Map{"Title": "Home"}), "partials/layout-chat")
		},
	).Name("page.chat")

	chat.Get(
		"/groups/:group_id",
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			channels, _ := model.GetGroupChannels(group.Id)
			subchannelsMap := make(map[uint64][]model.Channel)
			for _, ch := range channels {
				subchs, _ := model.GetSubchannels(ch.Id)
				subchannelsMap[ch.Id] = subchs
			}

			var activeChannel model.Channel
			if len(channels) > 0 {
				activeChannel = channels[0]
			}

			return htmx.TryRenderPage(ctx, "chat-group", MapPage(ctx, fiber.Map{
				"Title":          group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
			}), "partials/layout-chat")
		},
	).Name("page.group")

	chat.Get(
		"/groups/:group_id/channels/:channel_id",
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			channelId := fiber.Params[uint64](ctx, "channel_id")
			activeChannel, err := model.NewChannelFromId(channelId)
			if err != nil || !activeChannel.IsAvailable() {
				return ctx.Redirect().To(group.Url(ctx))
			}

			channels, _ := model.GetGroupChannels(group.Id)
			subchannelsMap := make(map[uint64][]model.Channel)
			for _, ch := range channels {
				subchs, _ := model.GetSubchannels(ch.Id)
				subchannelsMap[ch.Id] = subchs
			}

			return htmx.TryRenderPage(ctx, "chat-group", MapPage(ctx, fiber.Map{
				"Title":          activeChannel.Moniker + " - " + group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
			}), "partials/layout-chat")
		},
	).Name("page.group.channel")

	chat.Get(
		"/groups/join/:group_name",
		perms.GroupByName("group_name"),
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			if !group.IsAvailable() || group.Mode == model.GroupModePrivate {
				return htmx.AlertGroupNotFound
			}

			perms.MemberByAuthAndGroupNameFromCtx(ctx, "group_name")
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			if member.IsAvailable() {
				return ctx.Redirect().To(group.Url(ctx))
			}

			return htmx.TryRenderPage(ctx, "chat-group-join", MapPage(ctx, fiber.Map{"Title": "Join " + group.Moniker}), "partials/layout-chat")
		},
	).Name("page.group.join")

	return chat
}
