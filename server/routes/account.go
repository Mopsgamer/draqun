package routes

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type UserDelete struct {
	CurrentPassword model.Password `form:"current-password"`
	ConfirmUsername model.Name     `form:"confirm-username"`
}

type UserLogin struct {
	Email    model.Email    `form:"email"`
	Password model.Password `form:"password"`
}

type UserSignUp struct {
	*UserLogin
	Nickname        model.Moniker  `form:"nickname"`
	Username        model.Name     `form:"username"`
	Phone           model.Phone    `form:"phone"`
	ConfirmPassword model.Password `form:"confirm-password"`
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
			perms.UseBind[UserSignUp](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[UserSignUp](ctx, perms.LocalForm)

				u, _ := model.NewUserFromName(db, request.Username)
				if !u.IsEmpty() {
					return htmx.AlertUserExsistsNickname
				}

				u, _ = model.NewUserFromEmail(db, request.Email)
				if !u.IsEmpty() {
					return htmx.AlertUserExsistsEmail
				}

				if request.ConfirmPassword != request.Password {
					return htmx.AlertUserPasswordConfirm
				}

				hash, err := request.Password.Hash()
				if err != nil {
					return err
				}

				user := model.NewUser(db, request.Nickname, request.Username, request.Email, request.Phone, hash, "")
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

				user, err := model.NewUserFromEmail(db, request.Email)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
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

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))

					return ctx.SendStatus(fiber.StatusOK)
				}
				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/delete",
			perms.UserByAuth(db),
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
