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
	CurrentPassword string `form:"current-password"`
	ConfirmUsername string `form:"confirm-username"`
}

type UserLogin struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

type UserSignUp struct {
	*UserLogin
	Nickname        string `form:"nickname"`
	Username        string `form:"username"`
	Phone           string `form:"phone"`
	ConfirmPassword string `form:"confirm-password"`
}

func RouteAccount(app *fiber.App, db *model.DB) fiber.Router {
	group := app.Group("/account")
	routeAccountChange(group, db)
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
		Post("/create",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserSignUp](ctx, perms.LocalForm)

				if err := htmx.IsValidUserNick(request.Nickname); err != nil {
					return err
				}

				if err := htmx.IsValidUserName(request.Username); err != nil {
					return err
				}

				_, err := model.NewUserFromName(db, request.Username)
				if err != nil {
					return htmx.ErrUserExsistsNickname
				}

				_, err = model.NewUserFromEmail(db, request.Email)
				if err != nil {
					return htmx.ErrUserExsistsEmail
				}

				if err := htmx.IsValidUserPhone(request.Phone); err != nil {
					return err
				}

				// TODO: validate user avatar

				if request.ConfirmPassword != request.Password {
					return htmx.ErrUserPasswordConfirm
				}

				hash, err := model.HashPassword(request.Password)
				if err != nil {
					return err
				}

				user := model.NewUser(db, request.Nickname, request.Username, request.Email, request.Phone, hash, "")
				if err := user.Insert(); err != nil {
					return htmx.ErrDatabase.Join(err)
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
			perms.UseBind[UserSignUp](),
		).
		Post("/login",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserLogin](ctx, perms.LocalForm)
				if err := htmx.IsValidUserPassword(request.Password); err != nil {
					return err
				}

				if err := htmx.IsValidUserEmail(request.Email); err != nil {
					return err
				}

				user, err := model.NewUserFromEmail(db, request.Email)
				if err != nil {
					return htmx.ErrUserNotFound
				}

				if !user.CheckPassword(request.Password) {
					return htmx.ErrUserPassword
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
			perms.UseBind[UserLogin](),
		).
		Delete("/delete",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserDelete](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

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
				if err := user.Update(); err != nil {
					return htmx.ErrDatabase.Join(err)
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
			perms.UserByAuth(db),
			perms.UseBind[UserDelete](),
		)
}
