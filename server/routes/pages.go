package routes

import (
	"fmt"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func formatTenure(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	d := time.Since(t)
	if d < time.Minute {
		return "Just joined"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	months := days / 30
	if months < 12 {
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}
	years := months / 12
	remMonths := months % 12
	if remMonths == 0 {
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	}
	return fmt.Sprintf("%d yrs %d mos", years, remMonths)
}

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
		perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			return htmx.TryRenderPage(ctx, "group-settings", MapPage(ctx, fiber.Map{
				"Title":  "Group Settings - " + group.Moniker,
				"Group":  group,
				"Member": member,
				"Role":   role,
				"Tab":    "general",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/roles",
		perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			roles := group.Roles()
			return htmx.TryRenderPage(ctx, "group-roles", MapPage(ctx, fiber.Map{
				"Title":  "Group Roles - " + group.Moniker,
				"Group":  group,
				"Member": member,
				"Role":   role,
				"Roles":  roles,
				"Tab":    "roles",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/roles/create",
		perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			return htmx.TryRenderPage(ctx, "group-role-create", MapPage(ctx, fiber.Map{
				"Title":  "Create Role - " + group.Moniker,
				"Group":  group,
				"Member": member,
				"Role":   role,
				"Tab":    "roles",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/roles/:role_id",
		perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			roleId := fiber.Params[uint64](ctx, "role_id")
			targetRole, err := model.NewRoleFromId(roleId, group.Id)
			if err != nil || targetRole.IsEmpty() {
				return ctx.Redirect().To("/groups/" + fiber.Params[string](ctx, "group_id") + "/settings/roles")
			}
			return htmx.TryRenderPage(ctx, "group-role-edit", MapPage(ctx, fiber.Map{
				"Title":      "Edit Role - " + targetRole.Moniker,
				"Group":      group,
				"Member":     member,
				"Role":       role,
				"TargetRole": targetRole,
				"Tab":        "roles",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/members",
		perms.RequirePermLevel("group_id", model.PermNameMembers, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			users := group.UsersPage(1, 100)
			type MemberItem struct {
				User             model.User
				Member           model.Member
				Roles            []model.Role
				EffectiveRole    model.Role
				FirstJoinedAtStr string
				FirstTenureStr   string
				LastJoinedAtStr  string
				LastTenureStr    string
				AccountCreated   string
				IsOnline         bool
				MessagesCount    uint64
			}
			memberItems := make([]MemberItem, len(users))
			for i, u := range users {
				m, _ := model.NewMemberFromId(group.Id, u.Id)
				lastJoined := m.LastTimeJoinedAt()
				memberItems[i] = MemberItem{
					User:             u,
					Member:           m,
					Roles:            m.Roles(),
					EffectiveRole:    m.Role(),
					FirstJoinedAtStr: time.Time(m.FirstTimeJoinedAt).Format("2006-01-02"),
					FirstTenureStr:   formatTenure(time.Time(m.FirstTimeJoinedAt)),
					LastJoinedAtStr:  time.Time(lastJoined).Format("2006-01-02"),
					LastTenureStr:    formatTenure(time.Time(lastJoined)),
					AccountCreated:   time.Time(u.CreatedAt).Format("2006-01-02"),
					IsOnline:         u.IsOnline(),
					MessagesCount:    m.MessagesCount(),
				}
			}
			roles := group.Roles()
			activeBans := group.ActiveBans()
			type BanItem struct {
				Ban    model.ActionBan
				Target model.User
			}
			banItems := make([]BanItem, len(activeBans))
			for i, b := range activeBans {
				banItems[i] = BanItem{
					Ban:    b,
					Target: b.Target(),
				}
			}
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			callerLevel := model.GetUserPermissionLevel(group.Id, user.Id, 0, model.PermNameMembers)
			canKickBanGroup := (callerLevel == model.PermLevelAll || user.Id == group.OwnerId)

			return htmx.TryRenderPage(ctx, "group-members", MapPage(ctx, fiber.Map{
				"Title":           "Group Members - " + group.Moniker,
				"Group":           group,
				"Member":          member,
				"Role":            role,
				"MemberItems":     memberItems,
				"Roles":           roles,
				"BanItems":        banItems,
				"CanKickBanGroup": canKickBanGroup,
				"Tab":             "members",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/audit",
		perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			type AuditItem struct {
				Kind        string
				Target      model.User
				Creator     model.User
				Description string
				ActedAt     string
				Details     string
			}
			auditItems := []AuditItem{}
			for _, b := range group.Bans() {
				details := "Banned until " + time.Time(b.EndsAt).Format("2006-01-02 15:04")
				if b.RevokerId != 0 {
					details = "Unbanned by " + string(b.Revoker().Moniker)
				}
				auditItems = append(auditItems, AuditItem{
					Kind:        "Ban",
					Target:      b.Target(),
					Creator:     b.Creator(),
					Description: string(b.Description),
					ActedAt:     time.Time(b.ActedAt).Format("2006-01-02 15:04"),
					Details:     details,
				})
			}
			for _, k := range group.Kicks() {
				auditItems = append(auditItems, AuditItem{
					Kind:        "Kick",
					Target:      k.Target(),
					Creator:     k.Creator(),
					Description: string(k.Description),
					ActedAt:     time.Time(k.ActedAt).Format("2006-01-02 15:04"),
					Details:     "",
				})
			}
			for _, r := range group.RoleActions() {
				auditItems = append(auditItems, AuditItem{
					Kind:        "Role Action",
					Target:      model.User{},
					Creator:     r.Creator(),
					Description: string(r.Description),
					ActedAt:     time.Time(r.ActedAt).Format("2006-01-02 15:04"),
					Details:     "",
				})
			}
			for _, m := range group.Memberships() {
				actionStr := "Joined group"
				if !m.IsJoin {
					actionStr = "Left group"
				}
				auditItems = append(auditItems, AuditItem{
					Kind:        "Membership",
					Target:      m.User(),
					Creator:     model.User{},
					Description: actionStr,
					ActedAt:     time.Time(m.ActedAt).Format("2006-01-02 15:04"),
					Details:     "",
				})
			}
			return htmx.TryRenderPage(ctx, "group-audit", MapPage(ctx, fiber.Map{
				"Title":      "Audit Log - " + group.Moniker,
				"Group":      group,
				"Member":     member,
				"Role":       role,
				"AuditItems": auditItems,
				"Tab":        "audit",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/channels",
		perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			channels, _ := model.GetGroupChannels(group.Id)
			type ChannelItem struct {
				Channel     model.Channel
				Subchannels []model.Channel
			}
			channelItems := make([]ChannelItem, len(channels))
			for i, ch := range channels {
				var subchs []model.Channel
				if ch.Type != model.ChannelTypeTicket {
					subchs, _ = model.GetSubchannels(ch.Id)
				}
				channelItems[i] = ChannelItem{
					Channel:     ch,
					Subchannels: subchs,
				}
			}
			roles := group.Roles()
			return htmx.TryRenderPage(ctx, "group-channels", MapPage(ctx, fiber.Map{
				"Title":        "Group Channels - " + group.Moniker,
				"Group":        group,
				"Member":       member,
				"Role":         role,
				"ChannelItems": channelItems,
				"Channels":     channels,
				"Roles":        roles,
				"Tab":          "channels",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/permissions",
		perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			permissions, _ := model.GetGroupPermissions(group.Id)
			return htmx.TryRenderPage(ctx, "group-permissions", MapPage(ctx, fiber.Map{
				"Title":       "Group Permissions - " + group.Moniker,
				"Group":       group,
				"Member":      member,
				"Role":        role,
				"Permissions": permissions,
				"Tab":         "permissions",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/channels/create",
		perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			channels, _ := model.GetGroupChannels(group.Id)
			return htmx.TryRenderPage(ctx, "group-channel-create", MapPage(ctx, fiber.Map{
				"Title":    "Create Channel - " + group.Moniker,
				"Group":    group,
				"Member":   member,
				"Role":     role,
				"Channels": channels,
				"Tab":      "channels",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/channels/:channel_id",
		perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			channelId := fiber.Params[uint64](ctx, "channel_id")
			targetChannel, err := model.NewChannelFromId(channelId)
			if err != nil || !targetChannel.IsAvailable() {
				return ctx.Redirect().To(fmt.Sprintf("/groups/%d/settings/channels", group.Id))
			}

			// Verify channel belongs to this group
			chGroup, errG := model.GetChannelGroup(channelId)
			if errG != nil || chGroup == nil || chGroup.Id != group.Id {
				return ctx.Redirect().To(fmt.Sprintf("/groups/%d/settings/channels", group.Id))
			}

			channels, _ := model.GetGroupChannels(group.Id)
			var candidateParents []model.Channel
			for _, ch := range channels {
				if ch.Id != targetChannel.Id {
					candidateParents = append(candidateParents, ch)
				}
			}

			var currentParentId uint64
			if parent := targetChannel.ParentChannel(); parent != nil {
				currentParentId = parent.Id
			}

			return htmx.TryRenderPage(ctx, "group-channel-edit", MapPage(ctx, fiber.Map{
				"Title":           "Edit Channel - " + targetChannel.Moniker,
				"Group":           group,
				"Member":          member,
				"Role":            role,
				"TargetChannel":   targetChannel,
				"Channels":        candidateParents,
				"CurrentParentId": currentParentId,
				"Tab":             "channels",
			}), "partials/main")
		},
	)
	routePagesChat(app)
}
