package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type UserChangeName struct {
	NewNickname string `form:"new-nickname"`
	NewName     string `form:"new-username"`
}

type UserChangeEmail struct {
	CurrentPassword string `form:"current-password"`
	NewEmail        string `form:"new-email"`
}

type UserChangePhone struct {
	CurrentPassword string `form:"current-password"`
	NewPhone        string `form:"new-phone"`
}

type UserChangePassword struct {
	CurrentPassword string `form:"current-password"`
	NewPassword     string `form:"new-password"`
	ConfirmPassword string `form:"confirm-password"`
}

func routeAccountChange(router fiber.Router, db *model.DB) fiber.Router {
	return router.Group("/change").
		Put("/name",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangeName](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewNickname == user.Moniker && request.NewName == user.Name {
					return htmx.ErrUseless
				}

				if err := htmx.IsValidUserNick(request.NewNickname); err != nil {
					return err
				}

				if err := htmx.IsValidUserName(request.NewName); err != nil {
					return err
				}

				_, err := model.NewUserFromName(db, request.NewName)
				if err != nil {
					return htmx.ErrUserExsistsName
				}

				user.Moniker = request.NewNickname
				user.Name = request.NewName

				if err := user.Update(); err != nil {
					return htmx.ErrDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.UserByAuth(db),
			perms.UseBind[UserChangeName](),
		).
		Put("/email",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangeEmail](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewEmail == user.Email {
					return htmx.ErrUseless
				}

				if err := htmx.IsValidUserEmail(request.NewEmail); err != nil {
					return err
				}

				_, err := model.NewUserFromEmail(db, request.NewEmail)
				if err != nil {
					return htmx.ErrUserExsistsEmail
				}

				if err := htmx.IsValidUserPassword(request.CurrentPassword); err != nil {
					return err
				}

				if !user.CheckPassword(request.CurrentPassword) {
					return htmx.ErrUserPassword
				}

				user.Email = request.NewEmail

				if err := user.Update(); err != nil {
					return htmx.ErrDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.UserByAuth(db),
			perms.UseBind[UserChangeEmail](),
		).
		Put("/phone",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangePhone](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewPhone == user.Phone {
					return htmx.ErrUseless
				}

				if err := htmx.IsValidUserPhone(request.NewPhone); err != nil {
					return err
				}

				if !user.CheckPassword(request.CurrentPassword) {
					return htmx.ErrUserPassword
				}

				user.Phone = request.NewPhone

				if err := user.Update(); err != nil {
					return htmx.ErrDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.UserByAuth(db),
			perms.UseBind[UserChangePhone](),
		).
		Put("/password",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserChangePassword](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if request.NewPassword == user.Password {
					return htmx.ErrUseless
				}

				if err := htmx.IsValidUserPassword(request.CurrentPassword); err != nil {
					return err
				}

				if request.ConfirmPassword != request.NewPassword {
					return htmx.ErrUserPasswordConfirm
				}

				if !user.CheckPassword(request.CurrentPassword) {
					return htmx.ErrUserPassword
				}

				user.Password = request.NewPassword

				if err := user.Update(); err != nil {
					return htmx.ErrDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.UserByAuth(db),
			perms.UseBind[UserChangePassword](),
		)
}
