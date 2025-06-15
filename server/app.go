package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

var db, errDBLoad = database.InitDB()

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS) (*fiber.App, error) {
	if errDBLoad != nil {
		log.Error(errDBLoad)
		return nil, errDBLoad
	}

	engine := NewAppHtmlEngine(db, embedFS, "client/templates")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	Chain := controller.NewChainFactory(func(ctx fiber.Ctx) controller.Controller { return controller.Controller{Ctx: ctx, DB: *db} })

	UseStatic := func(dir string) fiber.Handler {
		if embedFS == nil {
			return static.New(dir, static.Config{Browse: true})
		}

		Fs, err := fs.Sub(embedFS, dir)
		if err != nil {
			log.Fatal(err)
		}

		return static.New("", static.Config{Browse: true, FS: Fs})
	}

	// static
	app.Get("/static*", UseStatic("client/static"))
	app.Get("/partials*", UseStatic("client/templates/partials"))

	// pages
	app.Get("/", Chain(func(ctl controller.Controller) error {
		return ctl.Ctx.Render("homepage", fiber.Map{"Title": "Homepage", "IsHomePage": true}, "partials/main")
	}))
	app.Get("/terms", Chain(func(ctl controller.Controller) error {
		return ctl.Ctx.Render("terms", fiber.Map{"Title": "Terms", "CenterContent": true}, "partials/main")
	}))
	app.Get("/privacy", Chain(func(ctl controller.Controller) error {
		return ctl.Ctx.Render("privacy", fiber.Map{"Title": "Privacy", "CenterContent": true}, "partials/main")
	}))
	app.Get("/acknowledgements", Chain(func(ctl controller.Controller) error {
		return ctl.Ctx.Render("acknowledgements", fiber.Map{"Title": "Acknowledgements"}, "partials/main")
	}))
	app.Get("/settings", Chain(
		controller.CheckAuth(),
		func(ctl controller.Controller) error {
			user := ctl.Ctx.Locals("user").(*model_database.User)
			if user == nil {
				return ctl.Ctx.Redirect().To("/")
			}

			return ctl.Ctx.Render("settings", fiber.Map{"Title": "Settings"}, "partials/main")
		},
	))
	app.Get("/chat", Chain(func(ctl controller.Controller) error {
		return ctl.Ctx.Render("chat", fiber.Map{"Title": "Home", "IsChatPage": true})
	}))
	app.Get("/chat/groups/:group_id", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")

		return Chain(
			controller.CheckAuthMember(false, groupId, func(role model_database.Role) bool {
				return true
			}),
			func(ctl controller.Controller) error {
				member := fiber.Locals[*model_database.Member](ctl.Ctx, controller.LocalMember)
				if member == nil {
					return ctx.Redirect().To("/chat")
				}

				group := fiber.Locals[*model_database.Group](ctl.Ctx, controller.LocalGroup)
				return ctx.Render("chat", fiber.Map{"Title": group.Nick, "IsChatPage": true})
			},
		)(ctx)
	})
	app.Get("/chat/groups/join/:group_name", func(ctx fiber.Ctx) error {
		groupName := fiber.Params[string](ctx, "group_name")
		group := db.GroupByName(groupName)

		return Chain(
			controller.CheckAuthMember(false, group.Id, func(role model_database.Role) bool {
				return true
			}),
			func(ctl controller.Controller) error {
				member := fiber.Locals[*model_database.Member](ctl.Ctx, controller.LocalMember)
				if group == nil {
					return ctx.Redirect().To("/chat")
				}

				if member != nil {
					return ctx.Redirect().To(controller.PathRedirectGroup(group.Id))
				}

				return ctx.Render("chat", fiber.Map{"Title": "Join " + group.Nick, "IsChatPage": true})
			},
		)(ctx)
	})

	// get
	app.Get("/groups/:group_id/messages/page/:messages_page", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")
		page := fiber.Params[uint64](ctx, "messages_page")

		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				return bool(role.ChatRead)
			}),
			func(ctl controller.Controller) error {
				const MessagesPagination uint64 = 5
				messageList := ctl.DB.MessageListPage(groupId, page, MessagesPagination)
				str, _ := ctl.RenderString("partials/chat-messages", ctl.MapPage(&fiber.Map{
					"GroupId":            groupId,
					"MessageList":        ctl.DB.CachedMessageList(messageList),
					"MessagesPageNext":   page + 1,
					"MessagesPagination": MessagesPagination,
				}))
				return ctl.Ctx.SendString(str)
			},
		)(ctx)
	})

	// post
	type UserLogin struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}
	type UserSignUp struct {
		*UserLogin
		Nickname        string  `form:"nickname"`
		Username        string  `form:"username"`
		Phone           *string `form:"phone"`
		ConfirmPassword string  `form:"confirm-password"`
	}
	app.Post("/account/create", Chain(
		controller.CheckBindForm(&UserSignUp{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserSignUp](ctl.Ctx, controller.LocalForm)

			if err := model_database.IsValidUserNick(request.Nickname); err != nil {
				return err
			}

			if err := model_database.IsValidUserName(request.Username); err != nil {
				return err
			}

			if ctl.DB.UserByUsername(request.Username) != nil {
				return controller.ErrUserExsistsNickname
			}

			if ctl.DB.UserByEmail(request.Email) != nil {
				return controller.ErrUserExsistsEmail
			}

			if err := model_database.IsValidUserPhone(request.Phone); err != nil {
				return err
			}

			// TODO: validate user avatar

			if request.ConfirmPassword != request.Password {
				return controller.ErrUserPasswordConfirm
			}

			hash, err := model_database.HashPassword(request.Password)
			if err != nil {
				return nil
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

			if ctl.DB.UserCreate(*user) == nil {
				return fiber.ErrInternalServerError
			}

			token, err := user.GenerateToken()
			if err != nil {
				return err
			}

			ctl.Ctx.Cookie(&fiber.Cookie{
				Name:    "Authorization",
				Value:   "Bearer " + token,
				Expires: time.Now().Add(environment.UserAuthTokenExpiration),
			})

			ctl.HTMXRedirect(ctl.HTMXCurrentPath())
			return nil
		},
	))
	var ErrUserNotExists = errors.New("user not exists")
	app.Post("/account/login", Chain(
		controller.CheckBindForm(&UserLogin{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserLogin](ctl.Ctx, controller.LocalForm)
			if err := model_database.IsValidUserPassword(request.Password); err != nil {
				return err
			}

			if err := model_database.IsValidUserEmail(request.Email); err != nil {
				return err
			}

			user := ctl.DB.UserByEmail(request.Email)
			if user == nil {
				return ErrUserNotExists
			}

			if !user.CheckPassword(request.Password) {
				return controller.ErrUserPassword
			}

			token, err := user.GenerateToken()
			if err != nil {
				return err
			}

			ctl.Ctx.Cookie(&fiber.Cookie{
				Name:    "Authorization",
				Value:   "Bearer " + token,
				Expires: time.Now().Add(environment.UserAuthTokenExpiration),
			})
			ctl.HTMXRedirect(ctl.HTMXCurrentPath())
			return nil
		},
	))
	type GroupCreate struct {
		Name        string  `form:"name"`
		Nick        string  `form:"nick"`
		Password    *string `form:"password"`
		Mode        string  `form:"mode"`
		Description string  `form:"description"`
		Avatar      string  `form:"avatar"`
	}
	app.Post("/groups/create", Chain(
		controller.CheckAuth(),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*GroupCreate](ctl.Ctx, controller.LocalForm)
			if ctl.DB.GroupByName(request.Name) != nil {
				return controller.ErrGroupNotFound
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

			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)
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
			groupId := ctl.DB.GroupCreate(*group)
			if groupId == nil {
				return fiber.ErrInternalServerError
			}

			group.Id = *groupId
			ctl.Ctx.Locals(controller.LocalGroup, group)

			member := model_database.Member{
				GroupId:  group.Id,
				UserId:   user.Id,
				Nick:     nil,
				IsOwner:  true,
				IsBanned: false,
			}

			if !ctl.DB.MemberCreate(member) {
				return fiber.ErrInternalServerError
			}

			right := model_database.RoleDefault
			rightId := ctl.DB.RoleCreate(right)
			if rightId == nil {
				return fiber.ErrInternalServerError
			}
			right.Id = *rightId

			role := model_database.RoleAssign{
				GroupId: group.Id,
				UserId:  user.Id,
				RightId: right.Id,
			}

			if !ctl.DB.RoleAssign(role) {
				return fiber.ErrInternalServerError
			}

			return nil
		},
		func(ctl controller.Controller) error {
			group := fiber.Locals[*model_database.Group](ctl.Ctx, controller.LocalGroup)
			ctl.HTMXRedirect(controller.PathRedirectGroup(group.Id))
			return nil
		}))
	type MessageCreate struct {
		Content string `form:"content"`
	}
	app.Post("/groups/:group_id/messages/create", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")
		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(member.IsOwner || (role.ChatRead && role.ChatWrite && !member.IsBanned))
			}),
			func(ctl controller.Controller) error {
				request := fiber.Locals[*MessageCreate](ctl.Ctx, controller.LocalForm)
				user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)
				group := fiber.Locals[*model_database.Group](ctl.Ctx, controller.LocalGroup)

				message := &model_database.Message{
					GroupId:   groupId,
					AuthorId:  user.Id,
					Content:   strings.TrimSpace(request.Content),
					CreatedAt: time.Now(),
				}

				if !model_database.IsValidMessageContent(message.Content) {
					return controller.ErrChatMessageContent
				}

				messageId := ctl.DB.MessageCreate(*message)
				if messageId == nil {
					return fiber.ErrInternalServerError
				}

				message.Id = *messageId
				str, _ := ctl.RenderString("partials/chat-messages", ctl.MapPage(&fiber.Map{
					"MessageList": ctl.DB.CachedMessageList([]model_database.Message{*message}),
				}))

				controller_ws.UserSessionMap.Push(
					func(userId uint64) bool {
						member := ctl.DB.MemberById(group.Id, userId)
						if member == nil {
							return false
						}

						if member.IsOwner {
							return true
						}

						role := ctl.DB.MemberRights(group.Id, userId)
						return bool(member.IsOwner || (role.ChatRead && !member.IsBanned))
					},
					controller.WrapOob("beforeend:#chat", &str),
					controller_ws.SubForMessages,
				)

				return ctl.Ctx.SendString("")
			},
		)(ctx)
	})

	// put
	type UserChangeName struct {
		NewNickname string `form:"new-nickname"`
		NewName     string `form:"new-username"`
	}
	app.Put("/account/change/name", Chain(
		controller.CheckAuth(),
		controller.CheckBindForm(&UserChangeName{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserChangeName](ctl.Ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)

			if request.NewNickname == user.Nick && request.NewName == user.Name {
				return controller.ErrUseless
			}

			if err := model_database.IsValidUserNick(request.NewNickname); err != nil {
				return err
			}

			if err := model_database.IsValidUserName(request.NewName); err != nil {
				return err
			}

			if ctl.DB.UserByUsername(request.NewName) != nil && request.NewNickname == user.Nick {
				return controller.ErrUserExsistsName
			}

			user.Nick = request.NewNickname
			user.Name = request.NewName

			if !ctl.DB.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			ctl.HTMXRefresh()
			return nil
		},
	))
	type UserChangeEmail struct {
		CurrentPassword string `form:"current-password"`
		NewEmail        string `form:"new-email"`
	}
	app.Put("/account/change/email", Chain(
		controller.CheckAuth(),
		controller.CheckBindForm(&UserChangeEmail{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserChangeEmail](ctl.Ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)

			if request.NewEmail == user.Email {
				return controller.ErrUseless
			}

			if err := model_database.IsValidUserEmail(request.NewEmail); err != nil {
				return err
			}

			if ctl.DB.UserByEmail(request.NewEmail) != nil {
				return controller.ErrUserExsistsEmail
			}

			if err := model_database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return controller.ErrUserPassword
			}

			user.Email = request.NewEmail

			if !ctl.DB.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			ctl.HTMXRefresh()
			return nil
		},
	))
	type UserChangePhone struct {
		CurrentPassword string  `form:"current-password"`
		NewPhone        *string `form:"new-phone"`
	}
	app.Put("/account/change/phone", Chain(
		controller.CheckAuth(),
		controller.CheckBindForm(&UserChangePhone{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserChangePhone](ctl.Ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)

			if request.NewPhone == user.Phone {
				return controller.ErrUseless
			}

			if err := model_database.IsValidUserPhone(request.NewPhone); err != nil {
				return err
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return controller.ErrUserPassword
			}

			user.Phone = request.NewPhone

			if !ctl.DB.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			ctl.HTMXRefresh()
			return nil
		},
	))
	type UserChangePassword struct {
		CurrentPassword string `form:"current-password"`
		NewPassword     string `form:"new-password"`
		ConfirmPassword string `form:"confirm-password"`
	}
	app.Put("/account/change/password", Chain(
		controller.CheckAuth(),
		controller.CheckBindForm(&UserChangePassword{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserChangePassword](ctl.Ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)

			if request.NewPassword == user.Password {
				return controller.ErrUseless
			}

			if err := model_database.IsValidUserPassword(request.CurrentPassword); err != nil {
				return err
			}

			if request.ConfirmPassword != request.NewPassword {
				return controller.ErrUserPasswordConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return controller.ErrUserPassword
			}

			user.Password = request.NewPassword

			if !ctl.DB.UserUpdate(*user) {
				return fiber.ErrInternalServerError
			}

			ctl.HTMXRefresh()
			return nil
		},
	))
	app.Put("/account/logout", Chain(
		func(ctl controller.Controller) error {
			ctl.Ctx.Cookie(&fiber.Cookie{
				Name:    "Authorization",
				Value:   "",
				Expires: time.Now(),
			})

			ctl.HTMXRedirect(ctl.HTMXCurrentPath())
			return ctl.Ctx.Render("partials/redirecting", fiber.Map{})
		},
	))
	app.Put("/groups/:group_id/join", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")

		return Chain(
			controller.CheckAuthMember(false, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(!member.IsBanned)
			}),
			func(ctl controller.Controller) error {
				user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)
				member := fiber.Locals[*model_database.Member](ctl.Ctx, controller.LocalMember)
				if member != nil {
					return controller.ErrUseless
				}

				member = &model_database.Member{
					GroupId:  groupId,
					UserId:   user.Id,
					Nick:     nil,
					IsOwner:  false,
					IsBanned: false,
				}

				if !ctl.DB.MemberCreate(*member) {
					return fiber.ErrInternalServerError
				}

				if len(ctl.DB.MemberRoleList(groupId, user.Id)) < 1 {
					right := model_database.RoleDefault
					rightId := ctl.DB.RoleCreate(right)
					if rightId == nil {
						return fiber.ErrInternalServerError
					}
					right.Id = *rightId

					role := model_database.RoleAssign{
						GroupId: groupId,
						UserId:  user.Id,
						RightId: right.Id,
					}

					if !ctl.DB.RoleAssign(role) {
						return fiber.ErrInternalServerError
					}
				}

				ctl.HTMXRedirect(controller.PathRedirectGroup(groupId))
				return nil
			},
		)(ctx)
	})
	type GroupChange struct {
		Name        string  `form:"name"`
		Nick        string  `form:"nick"`
		Password    *string `form:"password"`
		Mode        string  `form:"mode"`
		Description string  `form:"description"`
		Avatar      string  `form:"avatar"`
	}
	app.Put("/groups/:group_id/change", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")
		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(member.IsOwner || role.GroupChange)
			}),
			controller.CheckBindForm(&GroupChange{}),
			func(ctl controller.Controller) error {
				request := fiber.Locals[*GroupChange](ctl.Ctx, controller.LocalForm)
				group := fiber.Locals[*model_database.Group](ctl.Ctx, controller.LocalGroup)
				hasChanges := request.Nick != group.Nick ||
					group.Name != request.Name ||
					group.Description != request.Description ||
					group.Mode != model_database.GroupMode(request.Mode) ||
					group.Password != request.Password

				if !hasChanges {
					return controller.ErrUseless
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
				if !ctl.DB.GroupUpdate(*group) {
					return fiber.ErrInternalServerError
				}

				ctl.HTMXRefresh()
				return nil
			},
		)(ctx)
	})

	// delete
	app.Delete("/groups/:group_id/leave", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")

		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(!member.IsOwner)
			}),
			func(ctl controller.Controller) error {
				// TODO: do not leave if last owner and there are other non-owner members.
				// Ask for assigning new owner before leave.

				// TODO: delete group on leave if no other members.

				ctl.HTMXRedirect("/chat")
				return nil
			},
		)(ctx)
	})
	app.Delete("/groups/:group_id", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")

		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(member.IsOwner || role.GroupChange)
			}),
			func(ctl controller.Controller) error {
				if !ctl.DB.GroupDelete(groupId) {
					return fiber.ErrInternalServerError
				}

				ctl.HTMXRefresh()
				return nil
			},
		)(ctx)
	})
	type UserDelete struct {
		CurrentPassword string `form:"current-password"`
		ConfirmUsername string `form:"confirm-username"`
	}
	app.Delete("/account/delete", Chain(
		controller.CheckAuth(),
		controller.CheckBindForm(&UserDelete{}),
		func(ctl controller.Controller) error {
			request := fiber.Locals[*UserDelete](ctl.Ctx, controller.LocalForm)
			user := fiber.Locals[*model_database.User](ctl.Ctx, controller.LocalAuth)

			if user.Nick != request.ConfirmUsername {
				return controller.ErrUserNameConfirm
			}

			if !user.CheckPassword(request.CurrentPassword) {
				return controller.ErrUserPassword
			}

			userOwnGroups := ctl.DB.UserOwnGroupList(user.Id)
			if len(userOwnGroups) > 0 {
				return controller.ErrUserDeleteOwnerAccount
			}

			if !ctl.DB.UserDelete(user.Id) {
				return fiber.ErrInternalServerError
			}

			ctl.Ctx.Cookie(&fiber.Cookie{
				Name:    "Authorization",
				Value:   "",
				Expires: time.Now(),
			})

			ctl.HTMXRedirect(ctl.HTMXCurrentPath())
			return nil
		},
	))

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
	app.Get("/groups/:group_id", func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, "group_id")
		return Chain(
			controller.CheckAuthMember(true, groupId, func(role model_database.Role) bool {
				member := fiber.Locals[*model_database.Member](ctx, controller.LocalMember)
				return bool(member.IsOwner || role.ChatRead)
			}),
			func(ctlHttp controller.Controller) error {
				if !websocket.IsWebSocketUpgrade(ctlHttp.Ctx) {
					return ctlHttp.Ctx.Next()
				}

				// ctlWs MUST be created before websocket handler,
				// because some methods become unavailable after protocol upgraded.
				ctlWs := controller_ws.New(ctlHttp)

				user := fiber.Locals[*model_database.User](ctlHttp.Ctx, controller.LocalAuth)

				websocket.New(func(conn *websocket.Conn) {
					ctlWs.Conn = conn

					controller_ws.UserSessionMap.Connect(user.Id, ctlWs)
					defer controller_ws.UserSessionMap.Close(user.Id, ctlWs)
					for !ctlWs.Closed {
						messageType, message, err := ctlWs.Conn.ReadMessage()
						if err != nil {
							break
						}

						start := time.Now()
						ctlWs.MessageType = messageType
						ctlWs.Message = message
						err = ctlWs.Flush()

						logWS(start, err, ctlWs)

						if err != nil {
							break
						}
					}
					ctlWs.Closed = true
					ctlWs.Conn.Close()
				})(ctlHttp.Ctx)

				return nil
			},
		)(ctx)
	})

	app.Use(func(ctx fiber.Ctx) error {
		return ctx.Render(
			"partials/x",
			fiber.Map{
				"Title":         fmt.Sprintf("%d", fiber.StatusNotFound),
				"StatusCode":    fiber.StatusNotFound,
				"StatusMessage": fiber.ErrNotFound.Message,
				"CenterContent": true,
			},
			"partials/main",
		)
	})

	return app, nil
}
