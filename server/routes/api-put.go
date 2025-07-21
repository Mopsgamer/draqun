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

func RegisterPutRoutes(app *fiber.App, db *goqu.Database) {
	type UserChangeName struct {
		NewNickname string `form:"new-nickname"`
		NewName     string `form:"new-username"`
	}
	app.Put("/account/change/name",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangeName](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			if request.NewNickname == user.Moniker && request.NewName == user.Name {
				return environment.ErrUseless
			}

			if err := database.IsValidUserNick(request.NewNickname); err != nil {
				return err
			}

			if err := database.IsValidUserName(request.NewName); err != nil {
				return err
			}

			foundUser, _ := database.NewUserFromName(db, request.NewName)
			if foundUser {
				return environment.ErrUserExsistsName
			}

			user.Moniker = request.NewNickname
			user.Name = request.NewName

			if !user.Update() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&UserChangeName{}),
	)
	type UserChangeEmail struct {
		CurrentPassword string `form:"current-password"`
		NewEmail        string `form:"new-email"`
	}
	app.Put("/account/change/email",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangeEmail](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			if request.NewEmail == user.Email {
				return environment.ErrUseless
			}

			if err := database.IsValidUserEmail(request.NewEmail); err != nil {
				return err
			}

			foundUser, _ := database.NewUserFromEmail(db, request.NewEmail)
			if foundUser {
				return environment.ErrUserExsistsEmail
			}

			if err := database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Email = request.NewEmail

			if !user.Update() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&UserChangeEmail{}),
	)
	type UserChangePhone struct {
		CurrentPassword string `form:"current-password"`
		NewPhone        string `form:"new-phone"`
	}
	app.Put("/account/change/phone",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangePhone](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			if request.NewPhone == user.Phone {
				return environment.ErrUseless
			}

			if err := database.IsValidUserPhone(request.NewPhone); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Phone = request.NewPhone

			if !user.Update() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&UserChangePhone{}),
	)
	type UserChangePassword struct {
		CurrentPassword string `form:"current-password"`
		NewPassword     string `form:"new-password"`
		ConfirmPassword string `form:"confirm-password"`
	}
	app.Put("/account/change/password",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangePassword](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			if request.NewPassword == user.Password {
				return environment.ErrUseless
			}

			if err := database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if request.ConfirmPassword != request.NewPassword {
				return environment.ErrUserPasswordConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Password = request.NewPassword

			if !user.Update() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&UserChangePassword{}),
	)
	app.Put("/account/logout",
		func(ctx fiber.Ctx) error {
			ctx.Cookie(&fiber.Cookie{
				Name:    controller.AuthCookieKey,
				Value:   "",
				Expires: time.Now(),
			})

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, htmx.Path(ctx))
				return ctx.Render("partials/redirecting", fiber.Map{})
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)
	type GroupChange struct {
		Name        string `form:"name"`
		Nick        string `form:"nick"`
		Password    string `form:"password"`
		Mode        string `form:"mode"`
		Description string `form:"description"`
		Avatar      string `form:"avatar"`
	}
	app.Put("/groups/:group_id/change",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*GroupChange](ctx, controller.LocalForm)
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			hasChanges := request.Nick != group.Moniker ||
				group.Name != request.Name ||
				group.Description != request.Description ||
				group.Mode != database.GroupMode(request.Mode) ||
				group.Password != request.Password

			if !hasChanges {
				return environment.ErrUseless
			}

			if err := database.IsValidGroupName(request.Name); err != nil {
				return err
			}

			if err := database.IsValidGroupNick(request.Nick); err != nil {
				return err
			}

			if err := database.IsValidGroupPassword(request.Password); err != nil {
				return err
			}

			if err := database.IsValidGroupDescription(request.Description); err != nil {
				return err
			}

			if err := database.IsValidGroupMode(request.Mode); err != nil {
				return err
			}

			group.Moniker = request.Nick
			group.Name = request.Name
			group.Description = request.Description
			group.Mode = database.GroupMode(request.Mode)
			group.Password = request.Password
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
		controller.CheckBindForm(&GroupChange{}),
	)
}
