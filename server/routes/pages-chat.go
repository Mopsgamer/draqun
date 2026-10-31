package routes

import (
	"fmt"
	"sort"
	"strings"
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
			user, _ := perms.UserByAuthFromCtx(ctx)
			bind := fiber.Map{
				"Title":        "Home",
				"IsDraqunStaff": model.IsDraqunStaff(user),
			}
			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, bind), "partials/layout-chat")
		},
	).Name("page.chat")

	chat.Get(
		"/groups/:group_id",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			groupId := fiber.Params[uint64](ctx, "group_id")
			isStaff := model.IsDraqunStaff(user)

			if group, errG := model.NewGroupFromId(groupId); errG == nil && !group.IsEmpty() && group.IsBanned(user.Id) && !isStaff {
				return ctx.Redirect().To(group.UrlJoin(ctx))
			}

			group, errGroup := model.NewGroupFromId(groupId)
			if errGroup != nil || group.IsEmpty() || group.IsDeleted {
				return ctx.Redirect().To("/chat")
			}

			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil && !isStaff {
				return ctx.Redirect().To("/chat")
			}

			isAdmin, isMod := model.IsDraqunAdminOrMod(user)
			isGlobalAdmin := isAdmin || isMod
			allChannels, _ := model.GetGroupChannels(group.Id)
			var channels []model.Channel
			for _, ch := range allChannels {
				if isStaff || model.GetUserPermissionLevel(group.Id, user.Id, ch.Id, model.PermNameRead) != model.PermLevelDisable {
					channels = append(channels, ch)
				}
			}
			if isGlobalAdmin && group.Name == "draqun" {
				channels = append(channels, model.GetOtelVirtualChannel(), model.GetConfigVirtualChannel())
			}
			subchannelsMap := make(map[uint64][]model.Channel)
			for _, ch := range channels {
				if !ch.IsTicket() && !ch.IsOtel() && !ch.IsConfig() {
					subchs, _ := model.GetSubchannels(ch.Id)
					subchannelsMap[ch.Id] = subchs
				}
			}

			var activeChannel model.Channel
			if len(channels) > 0 {
				activeChannel = channels[0]
			}

			starredIdsMap := make(map[uint64]bool)
			if !user.IsEmpty() {
				for _, sc := range user.StarredChannels() {
					starredIdsMap[sc.Channel.Id] = true
				}
			}

			isStarred := false
			if !user.IsEmpty() && !activeChannel.IsEmpty() {
				isStarred = starredIdsMap[activeChannel.Id]
			}

			bind := fiber.Map{
				"Title":          group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
				"StarredIdsMap":  starredIdsMap,
				"IsStarred":      isStarred,
			}
			if activeChannel.IsWall() {
				bind["IsWall"] = true
				bind["WallAttachments"], _ = model.GetChannelAttachments(activeChannel.Id)
				bind["CanWriteDisabled"] = model.GetUserPermissionLevel(group.Id, user.Id, activeChannel.Id, model.PermNameWrite) == model.PermLevelDisable
			}
			enrichTicketContext(ctx, activeChannel, bind)

			return htmx.TryRenderPage(ctx, "chat-group", MapPage(ctx, bind), "partials/layout-chat")
		},
	).Name("page.group")

	chat.Get(
		"/groups/:group_id/channels/:channel_id",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			groupId := fiber.Params[uint64](ctx, "group_id")
			isStaff := model.IsDraqunStaff(user)

			if group, errG := model.NewGroupFromId(groupId); errG == nil && !group.IsEmpty() && group.IsBanned(user.Id) && !isStaff {
				return ctx.Redirect().To(group.UrlJoin(ctx))
			}

			group, errGroup := model.NewGroupFromId(groupId)
			if errGroup != nil || group.IsEmpty() || group.IsDeleted {
				return ctx.Redirect().To("/chat")
			}

			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil && !isStaff {
				return ctx.Redirect().To("/chat")
			}

			isAdmin, isMod := model.IsDraqunAdminOrMod(user)
			isGlobalAdmin := isAdmin || isMod
			channelId := fiber.Params[uint64](ctx, "channel_id")
			activeChannel, err := model.NewChannelFromId(channelId)
			if err != nil || (!activeChannel.IsAvailable() && activeChannel.Id != model.OtelVirtualChannelID && activeChannel.Id != model.ConfigVirtualChannelID) {
				return ctx.Redirect().To(group.Url(ctx))
			}

			if activeChannel.IsOtel() || activeChannel.IsConfig() {
				if !isGlobalAdmin {
					return ctx.Redirect().To(group.Url(ctx))
				}
			} else if !isStaff && model.GetUserPermissionLevel(group.Id, user.Id, activeChannel.Id, model.PermNameRead) == model.PermLevelDisable {
				return ctx.Redirect().To(group.Url(ctx))
			}

			allChannels, _ := model.GetGroupChannels(group.Id)
			var channels []model.Channel
			for _, ch := range allChannels {
				if isStaff || model.GetUserPermissionLevel(group.Id, user.Id, ch.Id, model.PermNameRead) != model.PermLevelDisable {
					channels = append(channels, ch)
				}
			}
			if isGlobalAdmin && group.Name == "draqun" {
				channels = append(channels, model.GetOtelVirtualChannel(), model.GetConfigVirtualChannel())
			}

			subchannelsMap := make(map[uint64][]model.Channel)
			for _, ch := range channels {
				if !ch.IsTicket() && !ch.IsOtel() && !ch.IsConfig() {
					subchs, _ := model.GetSubchannels(ch.Id)
					subchannelsMap[ch.Id] = subchs
				}
			}

			starredIdsMap := make(map[uint64]bool)
			if !user.IsEmpty() {
				for _, sc := range user.StarredChannels() {
					starredIdsMap[sc.Channel.Id] = true
				}
			}

			isStarred := false
			if !user.IsEmpty() && !activeChannel.IsEmpty() {
				isStarred = starredIdsMap[activeChannel.Id]
			}

			bind := fiber.Map{
				"Title":          activeChannel.Moniker + " - " + group.Moniker,
				"Group":          group,
				"Channels":       channels,
				"SubchannelsMap": subchannelsMap,
				"ActiveChannel":  activeChannel,
				"StarredIdsMap":  starredIdsMap,
				"IsStarred":      isStarred,
			}
			if activeChannel.IsOtel() {
				bind["IsOtel"] = true
				bind["OtelStats"] = model.GetOtelStats()
			} else if activeChannel.IsConfig() {
				bind["IsConfig"] = true
				bind["Config"] = model.GetServerConfig()
			} else if activeChannel.IsWall() {
				bind["IsWall"] = true
				bind["WallAttachments"], _ = model.GetChannelAttachments(activeChannel.Id)
				bind["CanWriteDisabled"] = model.GetUserPermissionLevel(group.Id, user.Id, activeChannel.Id, model.PermNameWrite) == model.PermLevelDisable
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

			user, _ := perms.UserByAuthFromCtx(ctx)
			isBanned := group.IsBanned(user.Id)
			var banReason string
			var banEndsAtStr string
			var banDurationStr string
			if isBanned {
				if user.IsSuspended {
					banReason = user.SuspensionReason
					banEndsAtStr = "Indefinite / Permanent"
					banDurationStr = "Indefinite / Permanent"
				} else {
					ban := group.GetActiveBan(user.Id)
					if ban != nil {
						banReason = string(ban.Description)
						tEnds := time.Time(ban.EndsAt)
						tActed := time.Time(ban.ActedAt)
						if !tEnds.IsZero() {
							banEndsAtStr = tEnds.Format("Jan 02, 2006 15:04")
							if !tActed.IsZero() && tEnds.After(tActed) {
								days := int(tEnds.Sub(tActed).Hours()/24 + 0.5)
								banDurationStr = fmt.Sprintf("%d days", days)
							} else {
								banDurationStr = "Indefinite / Permanent"
							}
						} else {
							banEndsAtStr = "Indefinite / Permanent"
							banDurationStr = "Indefinite / Permanent"
						}
					} else {
						banEndsAtStr = "Indefinite / Permanent"
						banDurationStr = "Indefinite / Permanent"
					}
				}
			} else {
				perms.MemberByAuthAndGroupNameFromCtx(ctx, "group_name")
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)
				if member.IsAvailable() {
					return ctx.Redirect().To(group.Url(ctx))
				}
			}

			bind := fiber.Map{
				"Title":          "Join " + group.Moniker,
				"Group":          group,
				"IsBanned":       isBanned,
				"BanReason":      banReason,
				"BanDurationStr": banDurationStr,
				"BanEndsAtStr":   banEndsAtStr,
			}

			return htmx.TryRenderPage(ctx, "chat-group-join", MapPage(ctx, bind), "partials/layout-chat")
		},
	).Name("page.group.join")

	chat.Get(
		"/admin/search",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if !model.IsDraqunStaff(user) {
				return ctx.Redirect().To("/chat")
			}

			q := strings.TrimSpace(ctx.Query("q", ""))
			filterType := ctx.Query("type", "all")
			if filterType != "groups" && filterType != "users" {
				filterType = "all"
			}

			bind := fiber.Map{
				"Title": "Global Admin Search",
				"Query": q,
				"Type":  filterType,
			}

			var groups []model.Group
			var users []model.User

			if filterType == "all" || filterType == "groups" {
				groups, _ = model.SearchAllGroups(q, 30)
			}
			if filterType == "all" || filterType == "users" {
				users, _ = model.SearchAllUsers(q, 30)
			}

			type UserWithDMs struct {
				model.User
				DirectChannels []model.UserDirectChannelPair
			}
			var userItems []UserWithDMs
			for _, u := range users {
				dps, _ := model.GetUserDirectChannelPairs(u.Id, 10)
				userItems = append(userItems, UserWithDMs{
					User:           u,
					DirectChannels: dps,
				})
			}

			bind["Groups"] = groups
			bind["Users"] = userItems

			return htmx.TryRenderPage(ctx, "admin-search", MapPage(ctx, bind), "partials/layout-chat")
		},
	).Name("page.admin.search")

	chat.Get(
		"/admin/search/results",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			if !model.IsDraqunStaff(user) {
				return ctx.Status(fiber.StatusForbidden).SendString("Unauthorized")
			}

			q := strings.TrimSpace(ctx.Query("q", ""))
			filterType := ctx.Query("type", "all")
			if filterType != "groups" && filterType != "users" {
				filterType = "all"
			}

			bind := fiber.Map{
				"Query": q,
				"Type":  filterType,
			}

			var groups []model.Group
			var users []model.User

			if filterType == "all" || filterType == "groups" {
				groups, _ = model.SearchAllGroups(q, 30)
			}
			if filterType == "all" || filterType == "users" {
				users, _ = model.SearchAllUsers(q, 30)
			}

			type UserWithDMs struct {
				model.User
				DirectChannels []model.UserDirectChannelPair
			}
			var userItems []UserWithDMs
			for _, u := range users {
				dps, _ := model.GetUserDirectChannelPairs(u.Id, 10)
				userItems = append(userItems, UserWithDMs{
					User:           u,
					DirectChannels: dps,
				})
			}

			bind["Groups"] = groups
			bind["Users"] = userItems

			return ctx.Render("partials/admin-search-results", bind)
		},
	).Name("page.admin.search.results")

	chat.Get(
		"/dm/:channel_id",
		func(ctx fiber.Ctx) error {
			user, _ := perms.UserByAuthFromCtx(ctx)
			channelId := fiber.Params[uint64](ctx, "channel_id")
			isStaff := model.IsDraqunStaff(user)

			if !isStaff && !model.IsUserInDirectChannel(channelId, user.Id) {
				return ctx.Redirect().To("/chat")
			}

			activeChannel, err := model.NewChannelFromId(channelId)
			if err != nil || !activeChannel.IsAvailable() {
				return ctx.Redirect().To("/chat")
			}

			u1, u2, errUsers := model.GetDirectChannelUsers(channelId)
			if errUsers != nil {
				return ctx.Redirect().To("/chat")
			}

			var otherUser model.User
			if user.Id == u1.Id {
				otherUser = u2
			} else if user.Id == u2.Id {
				otherUser = u1
			} else {
				otherUser = u1
			}

			bind := fiber.Map{
				"Title":         otherUser.Moniker,
				"ActiveChannel": activeChannel,
				"OtherUser":     otherUser,
				"IsDirectChat":  true,
			}

			return htmx.TryRenderPage(ctx, "chat", MapPage(ctx, bind), "partials/layout-chat")
		},
	).Name("page.dm")

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

	onlyParentsParam := ctx.FormValue("only_parents")
	if onlyParentsParam == "" {
		onlyParentsParam = ctx.Query("only_parents", "")
	}
	allParam := ctx.FormValue("all")
	if allParam == "" {
		allParam = ctx.Query("all", "")
	}

	var isOnlyParents bool
	if onlyParentsParam != "" {
		isOnlyParents = onlyParentsParam == "true" || onlyParentsParam == "1"
	} else if allParam != "" {
		isOnlyParents = !(allParam == "true" || allParam == "1")
	} else {
		isOnlyParents = false
	}
	bind["OnlyParentsFilter"] = isOnlyParents

	qQuery := ctx.FormValue("q")
	if qQuery == "" {
		qQuery = ctx.Query("q", "")
	}
	bind["SearchQuery"] = qQuery

	assignedParam := ctx.FormValue("assigned_to_me")
	if assignedParam == "" {
		assignedParam = ctx.Query("assigned_to_me", "")
	}
	isAssignedToMe := assignedParam == "true" || assignedParam == "1"
	bind["AssignedToMeFilter"] = isAssignedToMe

	createdByParam := ctx.FormValue("created_by_me")
	if createdByParam == "" {
		createdByParam = ctx.Query("created_by_me", "")
	}
	isCreatedByMe := createdByParam == "true" || createdByParam == "1"
	bind["CreatedByMeFilter"] = isCreatedByMe

	currentUser, _ := perms.UserByAuthFromCtx(ctx)

	if activeChannel.IsTicket() || (parentChannel != nil && parentChannel.IsTicket()) {
		if parentChannel == nil && activeChannel.IsTicket() {
			// Top-level ticket channel (Issues container)
			var candidates []model.Channel
			if isOnlyParents {
				candidates, _ = model.GetSubchannels(activeChannel.Id)
			} else {
				candidates, _ = model.GetAllDescendantTickets(activeChannel.Id)
			}

			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].Id > candidates[j].Id
			})

			openCount := 0
			closedCount := 0
			for _, sub := range candidates {
				if sub.Status == "closed" {
					closedCount++
				} else {
					openCount++
				}
			}

			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			groupId := group.Id
			if groupId == 0 {
				if g, errG := model.GetChannelGroup(activeChannel.Id); errG == nil && g != nil {
					groupId = g.Id
				}
			}

			qLower := strings.ToLower(strings.TrimSpace(qQuery))
			var filteredSubchs []model.Channel
			for _, sub := range candidates {
				if groupId != 0 && model.GetUserPermissionLevel(groupId, currentUser.Id, sub.Id, model.PermNameRead) == model.PermLevelDisable {
					continue
				}
				if statusFilter == "open" && sub.Status != "open" {
					continue
				}
				if statusFilter == "closed" && sub.Status != "closed" {
					continue
				}
				if isAssignedToMe {
					if currentUser.IsEmpty() || !model.IsChannelAssignee(sub.Id, currentUser.Id) {
						continue
					}
				}
				if isCreatedByMe {
					if currentUser.IsEmpty() || sub.AuthorId != currentUser.Id {
						continue
					}
				}
				if qLower != "" {
					idStr := fmt.Sprintf("%d", sub.Id)
					monikerLower := strings.ToLower(string(sub.Moniker))
					descLower := strings.ToLower(string(sub.Description))
					if !strings.Contains(monikerLower, qLower) && !strings.Contains(idStr, qLower) && !strings.Contains(descLower, qLower) {
						continue
					}
				}
				filteredSubchs = append(filteredSubchs, sub)
			}

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
			bind["OpenTicketsCount"] = openCount
			bind["ClosedTicketsCount"] = closedCount
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
