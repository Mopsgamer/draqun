package routes

import (
	"time"

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
		func(ctx fiber.Ctx) error {
			if err := perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id"); err != nil {
				return ctx.Redirect().To("/chat")
			}
			group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[model.Member](ctx, perms.LocalMember)
			role := fiber.Locals[model.Role](ctx, perms.LocalRights)
			users := group.UsersPage(1, 100)
			type MemberItem struct {
				User  model.User
				Roles []model.Role
			}
			memberItems := make([]MemberItem, len(users))
			for i, u := range users {
				m, _ := model.NewMemberFromId(group.Id, u.Id)
				memberItems[i] = MemberItem{
					User:  u,
					Roles: m.Roles(),
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
			return htmx.TryRenderPage(ctx, "group-members", MapPage(ctx, fiber.Map{
				"Title":       "Group Members - " + group.Moniker,
				"Group":       group,
				"Member":      member,
				"Role":        role,
				"MemberItems": memberItems,
				"Roles":       roles,
				"BanItems":    banItems,
				"Tab":         "members",
			}), "partials/main")
		},
	)
	app.Get(
		"/groups/:group_id/settings/audit",
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
					Details:     "Kicked from group",
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
					Creator:     m.User(),
					Description: actionStr,
					ActedAt:     time.Time(m.ActedAt).Format("2006-01-02 15:04"),
					Details:     actionStr,
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
	routePagesChat(app)
}
