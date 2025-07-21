package internal

import (
	_ "embed"
	"fmt"
	"io/fs"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS, clientEmbedded bool) (*fiber.App, error) {
	var db, errDBLoad = database.InitDB()
	if errDBLoad != nil {
		log.Error(errDBLoad)
		fmt.Println("Your database connection should be configured by DB_* variables: " + environment.GitHubRepo + "/blob/main/scripts/init.ts")
		return nil, errDBLoad
	}

	engine := NewAppHtmlEngine(db, embedFS, clientEmbedded, "client/templates")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			if htmx.IsHtmx(ctx) {
				return controller.HandleHTMXError(ctx, err)
			}
			return err
		},
	})

	app.Use(logger.New())

	static := controller.NewStaticFactory(embedFS, clientEmbedded)

	// static
	app.Get("/static", static(environment.StaticFolder))
	app.Get("/static/*", static(environment.StaticFolder))

	// pages
	app.Get("/",
		func(ctx fiber.Ctx) error {
			return ctx.Render("homepage", controller.MapPage(ctx, &fiber.Map{"Title": "Homepage", "IsHomePage": true}), "partials/main")
		},
	)
	app.Get("/terms",
		func(ctx fiber.Ctx) error {
			return ctx.Render("terms", controller.MapPage(ctx, &fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
	)
	app.Get("/privacy",
		func(ctx fiber.Ctx) error {
			return ctx.Render("privacy", controller.MapPage(ctx, &fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
	)
	app.Get("/acknowledgements",
		func(ctx fiber.Ctx) error {
			return ctx.Render("acknowledgements", controller.MapPage(ctx, &fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
	)
	app.Get("/settings",
		func(ctx fiber.Ctx) error {
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)
			if user.IsEmpty() {
				return ctx.Redirect().To("/")
			}

			return ctx.Render("settings", controller.MapPage(ctx, &fiber.Map{"Title": "Settings"}), "partials/main")
		},
	)
	app.Get("/chat",
		func(ctx fiber.Ctx) error {
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Home", "IsChatPage": true}))
		},
	)
	app.Get("/chat/groups/:group_id",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[database.Member](ctx, controller.LocalMember)
			if member.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": group.Moniker, "IsChatPage": true}))
		},
		controller.CheckGroupById(db, "group_id"),
	)
	app.Put("/groups/:group_id/join",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)
			member := fiber.Locals[database.Member](ctx, controller.LocalMember)
			if !member.IsEmpty() {
				return environment.ErrUseless
			}

			member = database.NewMemberEmpty(db, group.Id, user.Id)
			if !member.Insert() {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				// controller_ws.UserSessionMap.Push(
				// 	filter,
				// 	controller.WrapOob("beforeend:#chat", &str),
				// 	controller_ws.SubForMessages,
				// )

				htmx.Redirect(ctx, controller.PathRedirectGroup(group.Id))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckGroupById(db, "group_id"),
	)
	app.Get("/chat/groups/join/:group_name",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[database.Member](ctx, controller.LocalMember)
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			if group.IsEmpty() {
				return ctx.Redirect().To("/chat")
			}

			if member.IsEmpty() {
				return ctx.Redirect().To(controller.PathRedirectGroup(group.Id))
			}

			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Join " + group.Moniker, "IsChatPage": true}))
		},
		controller.CheckGroupByName(db, "group_name"),
	)

	// get
	app.Get("/groups/:group_id/messages/page/:messages_page",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			page := fiber.Params[uint](ctx, "messages_page")
			const MessagesPagination uint = 5
			messageList := group.MessagesPage(page, MessagesPagination)

			if htmx.IsHtmx(ctx) {
				bind := fiber.Map{
					"MessageList":        messageList,
					"MessagesPage":       page,
					"MessagesPagination": MessagesPagination,
				}

				return ctx.Render("partials/chat-messages", bind)
			}

			return ctx.JSON(messageList)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.CanReadMessages()
		}),
	)
	app.Get("/groups/:group_id/members/page/:members_page",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, controller.LocalGroup)
			page := fiber.Params[uint](ctx, "members_page")
			const MembersPagination uint = 5
			memberList := group.UsersPage(page, MembersPagination)

			if htmx.IsHtmx(ctx) {
				bind := fiber.Map{
					"Group":             group,
					"MemberList":        memberList,
					"MembersPage":       page,
					"MembersPagination": MembersPagination,
				}

				return ctx.Render("partials/chat-members", bind)
			}

			return ctx.JSON(memberList)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.CanDeleteMessages()
		}),
	)

	// post
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

			user := database.User{Db: db}
			if !user.FromEmail(request.Email) {
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

			user := database.User{Db: db}
			if user.FromName(request.Username) {
				return environment.ErrUserExsistsNickname
			}

			if user.FromEmail(request.Email) {
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

			user = database.User{
				Db: db,

				Moniker:    request.Nickname,
				Name:       request.Username,
				Email:      request.Email,
				Phone:      request.Phone,
				Password:   hash,
				CreatedAt:  time.Now(),
				LastSeenAt: time.Now(),
			}

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
			group := database.Group{Db: db}
			if !group.FromName(request.Name) {
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
			group.CreatorId = user.Id
			group.Moniker = request.Nick
			group.Name = request.Name
			group.Mode = database.GroupMode(request.Mode)
			group.Description = request.Description
			group.Password = request.Password
			group.Avatar = request.Avatar
			group.CreatedAt = time.Now()
			if !group.Insert() {
				return fiber.ErrInternalServerError
			}

			ctx.Locals(controller.LocalGroup, group)

			member := database.NewMemberEmpty(db, group.Id, user.Id)
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

	// put
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

			existingUser := database.User{Db: db}
			if existingUser.FromName(request.NewName) {
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

			existingUser := database.User{Db: db}
			if existingUser.FromEmail(request.NewEmail) {
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

	// delete
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

	logWS := func(start time.Time, err error, ws *controller_ws.ControllerWs) {
		colorErr := fiber.DefaultColors.Green
		if err != nil {
			colorErr = fiber.DefaultColors.Red
		}

		fmt.Printf(
			"%s | %s%3s%s | %13s | %15s | %d | %s%s%s \n",
			time.Now().Format("15:04:05"),
			colorErr,
			"ws",
			fiber.DefaultColors.Reset,
			time.Since(start),
			ws.IP,
			ws.MessageType,
			fiber.DefaultColors.Yellow,
			ws.Message,
			fiber.DefaultColors.Reset,
		)
	}

	// websoket
	app.Get("/groups/:group_id",
		func(ctx fiber.Ctx) error {
			if !websocket.IsWebSocketUpgrade(ctx) {
				return ctx.Next()
			}

			ctxWs := controller_ws.New(ctx)
			user := fiber.Locals[database.User](ctx, controller.LocalAuth)

			return websocket.New(func(conn *websocket.Conn) {
				ctxWs.Conn = conn
				controller_ws.UserSessionMap.Connect(user.Id, ctxWs)
				defer controller_ws.UserSessionMap.Close(user.Id, ctxWs)
				for !ctxWs.Closed {
					messageType, message, err := ctxWs.Conn.ReadMessage()
					if err != nil {
						break
					}

					start := time.Now()
					ctxWs.MessageType = messageType
					ctxWs.Message = message
					err = ctxWs.Flush()

					logWS(start, err, ctxWs)

					if err != nil {
						break
					}
				}
				ctxWs.Closed = true
				ctxWs.Conn.Close()
			})(ctx)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.CanReadMessages()
		}),
	)

	app.Use(func(ctx fiber.Ctx) error {
		if htmx.IsHtmx(ctx) {
			return ctx.Render(
				"partials/alert",
				fiber.Map{
					"Variant": "primary",
					"Message": "404",
				},
			)
		}
		if ctx.Method() == "GET" {
			return ctx.Render(
				"partials/x",
				fiber.Map{
					"Title":         "404",
					"StatusCode":    fiber.StatusNotFound,
					"StatusMessage": fiber.ErrNotFound.Message,
					"CenterContent": true,
				},
				"partials/main",
			)
		}

		return ctx.SendStatus(fiber.StatusNotFound)
	})

	return app, nil
}
