package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type UserChangeName struct {
	NewNickname model.Moniker `form:"new-nickname"`
	NewName     model.Name    `form:"new-username"`
}

type UserChangeEmail struct {
	CurrentPassword model.Password `form:"current-password"`
	NewEmail        model.Email    `form:"new-email"`
}

type UserChangePhone struct {
	CurrentPassword model.Password `form:"current-password"`
	NewPhone        model.Phone    `form:"new-phone"`
}

type UserChangePassword struct {
	CurrentPassword model.Password `form:"current-password"`
	NewPassword     model.Password `form:"new-password"`
	ConfirmPassword model.Password `form:"confirm-password"`
}

func routeAccountChange(router fiber.Router, db *model.DB) fiber.Router {
	return router.Group("/change").
		Put("/name",
			perms.UserByAuth(db),
			perms.UseBind[UserChangeName](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangeName](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewNickname == user.Moniker &&
					request.NewName == user.Name {
					return htmx.AlertUseless
				}

				_, err := model.NewUserFromName(db, request.NewName)
				if err != nil {
					return htmx.AlertUserExsistsName
				}

				user.Moniker = request.NewNickname
				user.Name = request.NewName
				if err := user.Validate(); err != nil {
					return err
				}

				if err := user.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/email",
			perms.UserByAuth(db),
			perms.UseBind[UserChangeEmail](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangeEmail](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewEmail == user.Email {
					return htmx.AlertUseless
				}

				_, err := model.NewUserFromEmail(db, request.NewEmail)
				if err != nil {
					return htmx.AlertUserExsistsEmail
				}

				if err := user.Password.Compare(request.CurrentPassword); err != nil {
					return htmx.AlertUserPassword.Join(err)
				}

				user.Email = request.NewEmail
				if err := user.Validate(); err != nil {
					return err
				}

				if err := user.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/phone",
			perms.UserByAuth(db),
			perms.UseBind[UserChangePhone](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangePhone](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewPhone == user.Phone {
					return htmx.AlertUseless
				}

				if err := user.Password.Compare(request.CurrentPassword); err != nil {
					return htmx.AlertUserPassword.Join(err)
				}

				user.Phone = request.NewPhone
				if err := user.Validate(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if err := user.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/password",
			perms.UserByAuth(db),
			perms.UseBind[UserChangePassword](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangePassword](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if err := user.Password.Compare(request.CurrentPassword); err != nil {
					return htmx.AlertUserPassword.Join(err)
				}

				if err := user.Password.Compare(request.NewPassword); err != nil {
					return htmx.AlertUseless.Join(err)
				}

				if request.ConfirmPassword != request.NewPassword {
					return htmx.AlertUserPasswordConfirm
				}

				var err error
				user.Password, err = request.NewPassword.Hash()
				if err != nil {
					return htmx.AlertEncryption.Join(err)
				}

				if err := user.Validate(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if err := user.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		)
}
