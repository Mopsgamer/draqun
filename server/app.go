package internal

import (
	_ "embed"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model_database"
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
		controller.PopulatePage(db),
	)
	app.Get("/terms",
		func(ctx fiber.Ctx) error {
			return ctx.Render("terms", controller.MapPage(ctx, &fiber.Map{"Title": "Terms", "CenterContent": true}), "partials/main")
		},
		controller.PopulatePage(db),
	)
	app.Get("/privacy",
		func(ctx fiber.Ctx) error {
			return ctx.Render("privacy", controller.MapPage(ctx, &fiber.Map{"Title": "Privacy", "CenterContent": true}), "partials/main")
		},
		controller.PopulatePage(db),
	)
	app.Get("/acknowledgements",
		func(ctx fiber.Ctx) error {
			return ctx.Render("acknowledgements", controller.MapPage(ctx, &fiber.Map{"Title": "Acknowledgements"}), "partials/main")
		},
		controller.PopulatePage(db),
	)
	app.Get("/settings",
		func(ctx fiber.Ctx) error {
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)
			if user == nil {
				return ctx.Redirect().To("/")
			}

			return ctx.Render("settings", controller.MapPage(ctx, &fiber.Map{"Title": "Settings"}), "partials/main")
		},
		controller.PopulatePage(db),
	)
	app.Get("/chat",
		func(ctx fiber.Ctx) error {
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Home", "IsChatPage": true}))
		},
		controller.PopulatePage(db),
	)
	app.Get("/chat/groups/:group_id",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			if member == nil {
				return ctx.Redirect().To("/chat")
			}

			group := fiber.Locals[*model_database.Group](ctx, controller.LocalGroup)
			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": group.Nick, "IsChatPage": true}))
		},
		controller.PopulatePage(db),
	)
	app.Get("/chat/groups/join/:group_name",
		func(ctx fiber.Ctx) error {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			group := fiber.Locals[*model_database.Group](ctx, controller.LocalGroup)
			if group == nil {
				return ctx.Redirect().To("/chat")
			}

			if member != nil {
				return ctx.Redirect().To(controller.PathRedirectGroup(group.Id))
			}

			return ctx.Render("chat", controller.MapPage(ctx, &fiber.Map{"Title": "Join " + group.Nick, "IsChatPage": true}))
		},
		controller.PopulatePage(db),
	)

	// get
	app.Get("/groups/:group_id/messages/page/:messages_page",
		func(ctx fiber.Ctx) error {
			groupId := fiber.Params[uint64](ctx, "group_id")
			page := fiber.Params[uint64](ctx, "messages_page")
			const MessagesPagination uint64 = 5
			messageList := db.MessageListPage(groupId, page, MessagesPagination)

			if htmx.IsHtmx(ctx) {
				bind := fiber.Map{
					"MessageList":        db.CachedMessageList(messageList),
					"MessagesPage":       page,
					"MessagesPagination": MessagesPagination,
				}

				return ctx.Render("partials/chat-messages", bind)
			}

			return ctx.JSON(messageList)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			return bool(role.ChatRead)
		}),
	)
	app.Get("/groups/:group_id/members/page/:members_page",
		func(ctx fiber.Ctx) error {
			groupId := fiber.Params[uint64](ctx, "group_id")
			page := fiber.Params[uint64](ctx, "members_page")
			group := fiber.Locals[*model_database.Group](ctx, controller.LocalGroup)
			const MembersPagination = 5
			memberList := db.MemberListPage(groupId, page, MembersPagination)

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
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			return bool(role.ChatRead)
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
			if err := model_database.IsValidUserPassword(request.Password); err != nil {
				return err
			}

			if err := model_database.IsValidUserEmail(request.Email); err != nil {
				return err
			}

			user := db.UserByEmail(request.Email)
			if user == nil {
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
		Nickname        string  `form:"nickname"`
		Username        string  `form:"username"`
		Phone           *string `form:"phone"`
		ConfirmPassword string  `form:"confirm-password"`
	}
	app.Post("/account/create",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserSignUp](ctx, controller.LocalForm)

			if err := model_database.IsValidUserNick(request.Nickname); err != nil {
				return err
			}

			if err := model_database.IsValidUserName(request.Username); err != nil {
				return err
			}

			if db.UserByUsername(request.Username) != nil {
				return environment.ErrUserExsistsNickname
			}

			if db.UserByEmail(request.Email) != nil {
				return environment.ErrUserExsistsEmail
			}

			if err := model_database.IsValidUserPhone(request.Phone); err != nil {
				return err
			}

			// TODO: validate user avatar

			if request.ConfirmPassword != request.Password {
				return environment.ErrUserPasswordConfirm
			}

			hash, err := model_database.HashPassword(request.Password)
			if err != nil {
				return err
			}

			user := &model_database.User{
				Nick:      request.Nickname,
				Name:      request.Username,
				Email:     request.Email,
				Phone:     request.Phone,
				Password:  hash,
				CreatedAt: time.Now(),
				LastSeen:  time.Now(),
			}

			if db.UserCreate(*user) == nil {
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
		Name        string  `form:"name"`
		Nick        string  `form:"nick"`
		Password    *string `form:"password"`
		Mode        string  `form:"mode"`
		Description string  `form:"description"`
		Avatar      string  `form:"avatar"`
	}
	app.Post("/groups/create",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*GroupCreate](ctx, controller.LocalForm)
			if db.GroupByName(request.Name) != nil {
				return environment.ErrGroupNotFound
			}

			if err := model_database.IsValidGroupName(request.Name); err != nil {
				return err
			}

			if err := model_database.IsValidGroupNick(request.Nick); err != nil {
				return err
			}

			if err := model_database.IsValidGroupPassword(request.Password); err != nil {
				return err
			}

			if err := model_database.IsValidGroupDescription(request.Description); err != nil {
				return err
			}

			if err := model_database.IsValidGroupMode(request.Mode); err != nil {
				return err
			}

			// TODO: validate group avatar

			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)
			group := &model_database.Group{
				CreatorId:   user.Id,
				Nick:        request.Nick,
				Name:        request.Name,
				Mode:        model_database.GroupMode(request.Mode),
				Description: request.Description,
				Password:    request.Password,
				Avatar:      request.Avatar,
				CreatedAt:   time.Now(),
			}
			groupId := db.GroupCreate(*group)
			if groupId == nil {
				return fiber.ErrInternalServerError
			}

			group.Id = *groupId
			ctx.Locals(controller.LocalGroup, group)

			member := model_database.Member{
				GroupId:  group.Id,
				UserId:   user.Id,
				Nick:     nil,
				IsOwner:  true,
				IsBanned: false,
			}

			if !db.MemberCreate(member) {
				return fiber.ErrInternalServerError
			}

			right := model_database.RoleDefault
			rightId := db.RoleCreate(right)
			if rightId == nil {
				return fiber.ErrInternalServerError
			}
			right.Id = *rightId

			role := model_database.RoleAssign{
				GroupId: group.Id,
				UserId:  user.Id,
				RightId: right.Id,
			}

			if !db.RoleAssign(role) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, controller.PathRedirectGroup(group.Id))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&GroupCreate{}),
		controller.CheckAuth(db),
	)
	type MessageCreate struct {
		Content string `form:"content"`
	}
	app.Post("/groups/:group_id/messages/create",
		func(ctx fiber.Ctx) error {
			groupId := fiber.Params[uint64](ctx, "group_id")
			request := fiber.Locals[*MessageCreate](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)
			group := fiber.Locals[*model_database.Group](ctx, controller.LocalGroup)

			message := &model_database.Message{
				GroupId:   groupId,
				AuthorId:  user.Id,
				Content:   strings.TrimSpace(request.Content),
				CreatedAt: time.Now(),
			}

			if err := model_database.IsValidMessageContent(message.Content); err != nil {
				return err
			}

			messageId := db.MessageCreate(*message)
			if messageId == nil {
				return fiber.ErrInternalServerError
			}

			message.Id = *messageId

			filter := func(userId uint64) bool {
				member := db.MemberById(group.Id, userId)
				if member == nil {
					return false
				}

				if member.IsOwner {
					return true
				}

				role := db.MemberRights(group.Id, userId)
				return bool(role.ChatRead && !member.IsBanned)
			}

			if htmx.IsHtmx(ctx) {
				buf, err := controller.RenderBuffer(app, "partials/chat-messages", &fiber.Map{
					"MessageList": db.CachedMessageList([]model_database.Message{*message}),
					"Group":       group,
					"User":        user,
				})
				if err != nil {
					return err
				}
				str := buf.String()

				controller_ws.UserSessionMap.Push(
					filter,
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
		controller.CheckBindForm(&MessageCreate{}),
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			return bool(member.IsOwner || (role.ChatRead && role.ChatWrite && !member.IsBanned))
		}),
	)

	// put
	type UserChangeName struct {
		NewNickname string `form:"new-nickname"`
		NewName     string `form:"new-username"`
	}
	app.Put("/account/change/name",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangeName](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

			if request.NewNickname == user.Nick && request.NewName == user.Name {
				return environment.ErrUseless
			}

			if err := model_database.IsValidUserNick(request.NewNickname); err != nil {
				return err
			}

			if err := model_database.IsValidUserName(request.NewName); err != nil {
				return err
			}

			if db.UserByUsername(request.NewName) != nil && request.NewNickname == user.Nick {
				return environment.ErrUserExsistsName
			}

			user.Nick = request.NewNickname
			user.Name = request.NewName

			if !db.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserChangeName{}),
		controller.CheckAuth(db),
	)
	type UserChangeEmail struct {
		CurrentPassword string `form:"current-password"`
		NewEmail        string `form:"new-email"`
	}
	app.Put("/account/change/email",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangeEmail](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

			if request.NewEmail == user.Email {
				return environment.ErrUseless
			}

			if err := model_database.IsValidUserEmail(request.NewEmail); err != nil {
				return err
			}

			if db.UserByEmail(request.NewEmail) != nil {
				return environment.ErrUserExsistsEmail
			}

			if err := model_database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Email = request.NewEmail

			if !db.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserChangeEmail{}),
		controller.CheckAuth(db),
	)
	type UserChangePhone struct {
		CurrentPassword string  `form:"current-password"`
		NewPhone        *string `form:"new-phone"`
	}
	app.Put("/account/change/phone",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangePhone](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

			if request.NewPhone == user.Phone {
				return environment.ErrUseless
			}

			if err := model_database.IsValidUserPhone(request.NewPhone); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Phone = request.NewPhone

			if !db.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserChangePhone{}),
		controller.CheckAuth(db),
	)
	type UserChangePassword struct {
		CurrentPassword string `form:"current-password"`
		NewPassword     string `form:"new-password"`
		ConfirmPassword string `form:"confirm-password"`
	}
	app.Put("/account/change/password",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserChangePassword](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

			if request.NewPassword == user.Password {
				return environment.ErrUseless
			}

			if err := model_database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if request.ConfirmPassword != request.NewPassword {
				return environment.ErrUserPasswordConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			user.Password = request.NewPassword

			if !db.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&UserChangePassword{}),
		controller.CheckAuth(db),
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
	app.Put("/groups/:group_id/join",
		func(ctx fiber.Ctx) error {
			groupId := fiber.Params[uint64](ctx, "group_id")

			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			if member != nil {
				return environment.ErrUseless
			}

			member = &model_database.Member{
				GroupId:  groupId,
				UserId:   user.Id,
				Nick:     nil,
				IsOwner:  false,
				IsBanned: false,
			}

			if !db.MemberCreate(*member) {
				return fiber.ErrInternalServerError
			}

			if len(db.MemberRoleList(groupId, user.Id)) < 1 {
				right := model_database.RoleDefault
				rightId := db.RoleCreate(right)
				if rightId == nil {
					return fiber.ErrInternalServerError
				}
				right.Id = *rightId

				role := model_database.RoleAssign{
					GroupId: groupId,
					UserId:  user.Id,
					RightId: right.Id,
				}

				if !db.RoleAssign(role) {
					return fiber.ErrInternalServerError
				}
			}

			if htmx.IsHtmx(ctx) {
				// controller_ws.UserSessionMap.Push(
				// 	filter,
				// 	controller.WrapOob("beforeend:#chat", &str),
				// 	controller_ws.SubForMessages,
				// )

				htmx.Redirect(ctx, controller.PathRedirectGroup(groupId))
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.PopulatePage(db),
	)
	type GroupChange struct {
		Name        string  `form:"name"`
		Nick        string  `form:"nick"`
		Password    *string `form:"password"`
		Mode        string  `form:"mode"`
		Description string  `form:"description"`
		Avatar      string  `form:"avatar"`
	}
	app.Put("/groups/:group_id/change",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*GroupChange](ctx, controller.LocalForm)
			group := fiber.Locals[*model_database.Group](ctx, controller.LocalGroup)
			hasChanges := request.Nick != group.Nick ||
				group.Name != request.Name ||
				group.Description != request.Description ||
				group.Mode != model_database.GroupMode(request.Mode) ||
				group.Password != request.Password

			if !hasChanges {
				return environment.ErrUseless
			}

			if err := model_database.IsValidGroupName(request.Name); err != nil {
				return err
			}

			if err := model_database.IsValidGroupNick(request.Nick); err != nil {
				return err
			}

			if err := model_database.IsValidGroupPassword(request.Password); err != nil {
				return err
			}

			if err := model_database.IsValidGroupDescription(request.Description); err != nil {
				return err
			}

			if err := model_database.IsValidGroupMode(request.Mode); err != nil {
				return err
			}

			group.Nick = request.Nick
			group.Name = request.Name
			group.Description = request.Description
			group.Mode = model_database.GroupMode(request.Mode)
			group.Password = request.Password
			if !db.GroupUpdate(*group) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckBindForm(&GroupChange{}),
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			return bool(member.IsOwner || role.GroupChange)
		}),
	)

	// delete
	app.Delete("/groups/:group_id/leave",
		func(ctx fiber.Ctx) error {
			// TODO: do not leave if last owner and there are other non-owner members.
			// Ask for assigning new owner before leave.

			// TODO: delete group on leave if no other members.

			if htmx.IsHtmx(ctx) {
				htmx.Redirect(ctx, "/chat")
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			return bool(!member.IsOwner)
		}),
	)
	app.Delete("/groups/:group_id",
		func(ctx fiber.Ctx) error {
			groupId := fiber.Params[uint64](ctx, "group_id")
			if !db.GroupDelete(groupId) {
				return fiber.ErrInternalServerError
			}

			if htmx.IsHtmx(ctx) {
				htmx.EnableRefresh(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			return bool(member.IsOwner || role.GroupChange)
		}),
	)
	type UserDelete struct {
		CurrentPassword string `form:"current-password"`
		ConfirmUsername string `form:"confirm-username"`
	}
	app.Delete("/account/delete",
		func(ctx fiber.Ctx) error {
			request := fiber.Locals[*UserDelete](ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

			if user.Nick != request.ConfirmUsername {
				return environment.ErrUserNameConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return environment.ErrUserPassword
			}

			userOwnGroups := db.UserOwnGroupList(user.Id)
			if len(userOwnGroups) > 0 {
				return environment.ErrUserDeleteOwnerAccount
			}

			if !db.UserDelete(user.Id) {
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
		controller.CheckBindForm(&UserDelete{}),
		controller.CheckAuth(db),
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
			user := fiber.Locals[*model_database.User](ctx, controller.LocalAuth)

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
		controller.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role model_database.Role) bool {
			member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
			return bool(member.IsOwner || role.ChatRead)
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
