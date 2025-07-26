package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RegisterDeleteRoutes(app *fiber.App, db *database.DB) {
	app.Delete("/groups/:group_id/leave",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
			member := fiber.Locals[database.Member](ctx, perms.LocalMember)

			isAlone := group.MembersCount() == 1
			if group.OwnerId == member.UserId && !isAlone {
				return htmx.ErrGroupMemberIsOnlyOwner
			}

			if isAlone {
				group.IsDeleted = true
				group.Update()
			}

			member.IsDeleted = true
			if !member.Update() {
				return htmx.ErrDatabase
			}
			if !member.LeaveActed() {
				return htmx.ErrDatabase
			}

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, "/chat")
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return true
		}),
	)
	app.Delete("/groups/:group_id",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
			group.IsDeleted = true
			if !group.Update() {
				return htmx.ErrDatabase
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.PermGroupChange.Has()
		}),
	)
	type UserDelete struct {
		CurrentPassword string `form:"current-password"`
		ConfirmUsername string `form:"confirm-username"`
	}
	app.Delete("/account/delete",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserDelete](ctx, perms.LocalForm)
			user := fiber.Locals[database.User](ctx, perms.LocalAuth)

			if user.Moniker != request.ConfirmUsername {
				return htmx.ErrUserNameConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return htmx.ErrUserPassword
			}

			if len(user.GroupListOwner()) > 0 {
				return htmx.ErrUserDeleteOwnerAccount
			}

			user.IsDeleted = true
			if !user.Update() {
				return htmx.ErrDatabase
			}

			ctx.Cookie(&fiber.Cookie{
				Name:    perms.AuthCookieKey,
				Value:   "",
				Expires: time.Now(),
			})

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, htmx.Path(ctx))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		perms.UserByAuth(db),
		perms.UseForm(&UserDelete{}),
	)
}
