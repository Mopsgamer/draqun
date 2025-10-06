package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type UserDelete struct {
	CurrentPassword model.Password `form:"current-password"`
	ConfirmUsername model.Name     `form:"confirm-name"`
}

type UserLogin struct {
	Email    model.Email    `form:"email"`
	Password model.Password `form:"password"`
}

type UserSignUp struct {
	UserLogin
	Moniker         model.Moniker  `form:"moniker"`
	Name            model.Name     `form:"name"`
	Phone           model.Phone    `form:"phone"`
	ConfirmPassword model.Password `form:"confirm-password"`
}

func RouteAccount(app *fiber.App) fiber.Router {
	group := app.Group("/account")
	routeAccountChange(group)
	return group.
		Put("/logout",
			func(ctx fiber.Ctx) error {
				ctx.Cookie(&fiber.Cookie{
					Name:    fiber.HeaderAuthorization,
					Value:   "",
					Expires: time.Now(),
				})

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))
					return ctx.Render("partials/redirecting", fiber.Map{})
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/",
			perms.UseBind[UserSignUp](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserSignUp](ctx, perms.LocalForm)

				existingUser, _ := model.NewUserFromName(request.Name)
				if !existingUser.IsEmpty() {
					return htmx.AlertUserExistsNickname
				}

				existingUser, _ = model.NewUserFromEmail(request.Email)
				if !existingUser.IsEmpty() {
					return htmx.AlertUserExistsEmail
				}

				if request.ConfirmPassword != request.Password {
					return htmx.AlertUserPasswordConfirm
				}

				hash, err := request.Password.Hash()
				if err != nil {
					return err
				}

				user := model.NewUser(request.Moniker, request.Name, request.Email, request.Phone, hash, "")
				if err := user.Validate(); err != nil {
					return err
				}

				if err := user.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				token, err := user.GenerateToken()
				if err != nil {
					return err
				}

				ctx.Cookie(&fiber.Cookie{
					Name:    fiber.HeaderAuthorization,
					Value:   token,
					Expires: time.Now().Add(environment.UserAuthTokenExpiration),
				})

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/login",
			perms.UseBind[UserLogin](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserLogin](ctx, perms.LocalForm)

				user, err := model.NewUserFromEmail(request.Email)
				if err != nil {
					if user.IsEmpty() {
						return htmx.AlertUserNotFound
					}
					return htmx.AlertDatabase.Join(err)
				}

				if err := user.Validate(); err != nil {
					return err
				}

				if user.Password.Compare(request.Password) != nil {
					return htmx.AlertUserPassword
				}

				token, err := user.GenerateToken()
				if err != nil {
					return err
				}

				ctx.Cookie(&fiber.Cookie{
					Name:    fiber.HeaderAuthorization,
					Value:   token,
					Expires: time.Now().Add(environment.UserAuthTokenExpiration),
				})

				if user.IsDeleted {
					user.IsDeleted = false
					if err := user.Update(); err != nil {
						return err
					}
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))

					return ctx.SendStatus(fiber.StatusOK)
				}
				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/",
			perms.UserByAuth(),
			perms.UseBind[UserDelete](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserDelete](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				if user.Name != request.ConfirmUsername {
					return htmx.AlertUserNameConfirm
				}

				if user.Password.Compare(request.CurrentPassword) != nil {
					return htmx.AlertUserPassword
				}

				if len(user.GroupListOwner()) > 0 {
					return htmx.AlertUserDeleteOwnerAccount
				}

				user.IsDeleted = true
				if err := user.Validate(); err != nil {
					return err
				}

				if err := user.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				ctx.Cookie(&fiber.Cookie{
					Name:    fiber.HeaderAuthorization,
					Value:   "",
					Expires: time.Now(),
				})

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		)
}
