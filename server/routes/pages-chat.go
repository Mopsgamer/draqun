package routes

import (
	"time"

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
				if !ch.IsTicket() {
					subchs, _ := model.GetSubchannels(ch.Id)
					subchannelsMap[ch.Id] = subchs
				}
			}

			var activeChannel model.Channel
			if len(channels) > 0 {
				activeChannel = channels[0]
			}

			bind := fiber.Map{
				"Title":          group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
			}
			enrichTicketContext(ctx, activeChannel, bind)

			return htmx.TryRenderPage(ctx, "chat-group", MapPage(ctx, bind), "partials/layout-chat")
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
				if !ch.IsTicket() {
					subchs, _ := model.GetSubchannels(ch.Id)
					subchannelsMap[ch.Id] = subchs
				}
			}

			bind := fiber.Map{
				"Title":          activeChannel.Moniker + " - " + group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
			}
			enrichTicketContext(ctx, activeChannel, bind)

			return htmx.TryRenderPage(ctx, "chat-group", MapPage(ctx, bind), "partials/layout-chat")
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

type TicketItem struct {
	Channel         model.Channel
	Author          model.User
	Assignees       []model.User
	Progress        model.TicketProgress
	CommentCount    int
	DuplicateTarget *model.Channel
}

func enrichTicketContext(ctx fiber.Ctx, activeChannel model.Channel, bind fiber.Map) {
	if activeChannel.IsEmpty() {
		return
	}

	parentChannels := activeChannel.ParentChannels()
	var parentChannel *model.Channel
	if len(parentChannels) > 0 {
		parentChannel = &parentChannels[0]
	}
	bind["ParentChannel"] = parentChannel
	bind["ParentChannels"] = parentChannels

	statusFilter := ctx.FormValue("status")
	if statusFilter == "" {
		statusFilter = ctx.Query("status", "open")
	}
	if statusFilter != "open" && statusFilter != "closed" && statusFilter != "all" {
		statusFilter = "open"
	}
	bind["StatusFilter"] = statusFilter

	if activeChannel.IsTicket() || (parentChannel != nil && parentChannel.IsTicket()) {
		if parentChannel == nil && activeChannel.IsTicket() {
			// Top-level ticket channel (Issues container)
			openSubchs, _ := model.GetTicketSubchannels(activeChannel.Id, "open")
			closedSubchs, _ := model.GetTicketSubchannels(activeChannel.Id, "closed")
			filteredSubchs, _ := model.GetTicketSubchannels(activeChannel.Id, statusFilter)

			var ticketItems []TicketItem
			for _, sub := range filteredSubchs {
				ticketItems = append(ticketItems, TicketItem{
					Channel:         sub,
					Author:          sub.AuthorUser(),
					Assignees:       sub.Assignees(),
					Progress:        sub.GetTicketProgress(),
					CommentCount:    sub.CommentCount(),
					DuplicateTarget: sub.DuplicateChannel(),
				})
			}

			bind["IsTicketContainer"] = true
			bind["OpenTicketsCount"] = len(openSubchs)
			bind["ClosedTicketsCount"] = len(closedSubchs)
			bind["TicketItems"] = ticketItems
		} else {
			// Ticket subchannel (individual issue view)
			subTickets, _ := model.GetSubchannels(activeChannel.Id)
			var subTicketItems []TicketItem
			for _, sub := range subTickets {
				subTicketItems = append(subTicketItems, TicketItem{
					Channel:         sub,
					Author:          sub.AuthorUser(),
					Assignees:       sub.Assignees(),
					Progress:        sub.GetTicketProgress(),
					CommentCount:    sub.CommentCount(),
					DuplicateTarget: sub.DuplicateChannel(),
				})
			}

			parentIdsMap := make(map[uint64]bool)
			for _, p := range parentChannels {
				parentIdsMap[p.Id] = true
			}

			bind["IsTicketDetail"] = true
			bind["TicketAuthor"] = activeChannel.AuthorUser()
			bind["TicketAssignees"] = activeChannel.Assignees()
			bind["TicketProgress"] = activeChannel.GetTicketProgress()
			bind["TicketDuplicateTarget"] = activeChannel.DuplicateChannel()
			bind["SubTickets"] = subTicketItems
			bind["ParentIdsMap"] = parentIdsMap

			if !time.Time(activeChannel.UpdatedAt).IsZero() {
				bind["TicketEditDateStr"] = time.Time(activeChannel.UpdatedAt).Format("Jan 02, 2006 15:04")
			}
		}
	}
}
