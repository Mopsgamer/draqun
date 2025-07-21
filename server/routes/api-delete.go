package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

func RegisterDeleteRoutes(app *fiber.App, db *goqu.Database) {
	app.Delete("/groups/:group_id/leave",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			isAlone := group.MembersCount() == 1
			if group.AdminsCount() == 1 && !isAlone {
				return environment.ErrGroupMemberIsOnlyAdmin
			}

			if isAlone {
				group.IsDeleted = true
				group.Update()
			}

			member := fiber.Locals[database.Member](ctx, controller.LocalMember)
			member.IsDeleted = true
			if !member.Update() {
				return fiber.ErrInternalServerError
			}
			if !member.LeaveActed() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, "/chat")
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return true
		}),
	)
	app.Delete("/groups/:group_id",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			group.IsDeleted = true
			if !group.Update() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.PermGroupChange.Has()
		}),
	)
	type UserDelete struct {
		CurrentPassword string `form:"current-password"`
		ConfirmUsername string `form:"confirm-username"`
	}
	app.Delete("/account/delete",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserDelete](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			if user.Moniker != request.ConfirmUsername {
				return environment.ErrUserNameConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			userOwnGroups := user.Groups()
			if len(userOwnGroups) > 0 {
				return environment.ErrUserDeleteOwnerAccount
			}

			user.IsDeleted = true
			if !user.Update() {
				return fiber.ErrInternalServerError
			}

			ctx.Cookie(&fiber.Cookie{
				Name:    controller.AuthCookieKey,
				Value:   "",
				Expires: time.Now(),
			})

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, htmx.Path(ctx))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&UserDelete{}),
	)
}
