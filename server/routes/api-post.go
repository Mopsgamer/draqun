package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

func RegisterPostRoutes(app *fiber.App, db *goqu.Database) {
	type UserLogin struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}
	app.Post("/account/login",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserLogin](ctx, controller.LocalForm)
			if err := database.IsValidUserPassword(request.Password); err != nil {
				return err
			}

			if err := database.IsValidUserEmail(request.Email); err != nil {
				return err
			}

			userFound, user := database.NewUserFromEmail(db, request.Email)
			if !userFound {
				return environment.ErrUserNotFound
			}

			if !user.CheckPassword(request.Password) {
				return environment.ErrUserPassword
			}

			token, err := user.GenerateToken()
			if err != nil {
				return err
			}

			ctx.Cookie(&fiber.Cookie{
				Name:    controller.AuthCookieKey,
				Value:   token,
				Expires: time.Now().Add(environment.UserAuthTokenExpiration),
			})

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, htmx.Path(ctx))

				return ctx.SendStatus(fiber.StatusOK)
			}
			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserLogin{}),
	)
	type UserSignUp struct {
		*UserLogin
		Nickname        string `form:"nickname"`
		Username        string `form:"username"`
		Phone           string `form:"phone"`
		ConfirmPassword string `form:"confirm-password"`
	}
	app.Post("/account/create",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserSignUp](ctx, controller.LocalForm)

			if err := database.IsValidUserNick(request.Nickname); err != nil {
				return err
			}

			if err := database.IsValidUserName(request.Username); err != nil {
				return err
			}

			userFound, _ := database.NewUserFromName(db, request.Username)
			if userFound {
				return environment.ErrUserExsistsNickname
			}

			userFound, _ = database.NewUserFromEmail(db, request.Email)
			if userFound {
				return environment.ErrUserExsistsEmail
			}

			if err := database.IsValidUserPhone(request.Phone); err != nil {
				return err
			}

			// TODO: validate user avatar

			if request.ConfirmPassword != request.Password {
				return environment.ErrUserPasswordConfirm
			}

			hash, err := database.HashPassword(request.Password)
			if err != nil {
				return err
			}

			user := database.NewUser(db, request.Nickname, request.Username, request.Email, request.Phone, hash, "")
			if !user.Insert() {
				return fiber.ErrInternalServerError
			}

			token, err := user.GenerateToken()
			if err != nil {
				return err
			}

			ctx.Cookie(&fiber.Cookie{
				Name:    controller.AuthCookieKey,
				Value:   token,
				Expires: time.Now().Add(environment.UserAuthTokenExpiration),
			})

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, htmx.Path(ctx))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserSignUp{}),
	)
	type GroupCreate struct {
		Name        string `form:"name"`
		Nick        string `form:"nick"`
		Password    string `form:"password"`
		Mode        string `form:"mode"`
		Description string `form:"description"`
		Avatar      string `form:"avatar"`
	}
	app.Post("/groups/create",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*GroupCreate](ctx, controller.LocalForm)
			groupFound, _ := database.NewGroupFromName(db, request.Name)
			if !groupFound {
				return environment.ErrGroupNotFound
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

			// TODO: validate group avatar

			user := fiber.Locals[database.User](ctx, controller.LocalAuth)
			group := database.NewGroup(db, user.Id, request.Nick, request.Name, database.GroupMode(request.Mode), request.Password, request.Description, request.Avatar)
			if !group.Insert() {
				return fiber.ErrInternalServerError
			}

			ctx.Locals(controller.LocalGroup, group)

			member := database.NewMember(db, group.Id, user.Id, "")
			if !member.Insert() {
				return fiber.ErrInternalServerError
			}

			everyone := database.NewRoleEveryone(db, group.Id)
			if !everyone.Insert() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, controller.PathRedirectGroup(group.Id))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuth(db),
		controller.CheckBindForm(&GroupCreate{}),
	)
	type MessageCreate struct {
		Content string `form:"content"`
	}
	app.Post("/groups/:group_id/messages/create",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*MessageCreate](ctx, controller.LocalForm)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)

			message := database.NewMessageFilled(db, group.Id, user.Id, request.Content)
			if err := database.IsValidMessageContent(message.Content); err != nil {
				return err
			}

			if !message.Insert() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				buf, err := controller.RenderBuffer(app, "partials/chat-messages", &fiber.Map{
					"MessageList": []database.Message{message},
					"Group":       group,
					"User":        user,
				})
				if err != nil {
					return err
				}
				str := buf.String()

				controller_ws.UserSessionMap.Push(
					controller.WrapOob("beforeend:#chat", &str),
					controller_ws.SubForMessages,
				)

				return ctx.SendStatus(fiber.StatusOK)
			}

			// controller_ws.UserSessionMap.Push(
			// 		filter,
			// 		...,
			// 		controller_ws.SubForMessages,
			// 	)

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.CanWriteMessages()
		}),
		controller.CheckBindForm(&MessageCreate{}),
	)
}
