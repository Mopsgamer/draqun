package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/Mopsgamer/draqun/websocket"
	"github.com/gofiber/fiber/v3"
)

func RouteGroups(app *fiber.App, db *model.DB) fiber.Router {
	type GroupCreate struct {
		Name        string `form:"name"`
		Nick        string `form:"nick"`
		Password    string `form:"password"`
		Mode        string `form:"mode"`
		Description string `form:"description"`
		Avatar      string `form:"avatar"`
	}
	type MessageCreate struct {
		Content string `form:"content"`
	}
	type GroupChange struct {
		Name        string `form:"name"`
		Nick        string `form:"nick"`
		Password    string `form:"password"`
		Mode        string `form:"mode"`
		Description string `form:"description"`
		Avatar      string `form:"avatar"`
	}
	return app.Group("/groups").
		Put("/:group_id/join",
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)
				if member.IsEmpty() {
					// never been a member
					member = model.NewMember(db, group.Id, member.UserId, "")
					if !member.Insert() {
						return htmx.ErrDatabase
					}
				} else if member.IsDeleted {
					// been a member, now isn't
					member.IsDeleted = false
					if !member.Update() {
						return htmx.ErrDatabase
					}
				} else {
					// already a member
					return htmx.ErrUseless
				}

				if !member.JoinActed() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					// controller_ws.UserSessionMap.Push(
					// 	filter,
					// 	controller.WrapOob("beforeend:#chat", &str),
					// 	controller_ws.SubForMessages,
					// )

					htmx.Redirect(ctx, group.Url(ctx))
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.GroupById(db, "group_id"),
		).
		Put("/:group_id/change",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[*GroupChange](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				hasChanges := request.Nick != group.Moniker ||
					group.Name != request.Name ||
					group.Description != request.Description ||
					group.Mode != model.GroupMode(request.Mode) ||
					group.Password != request.Password

				if !hasChanges {
					return htmx.ErrUseless
				}

				if err := htmx.IsValidGroupName(request.Name); err != nil {
					return err
				}

				if err := htmx.IsValidGroupNick(request.Nick); err != nil {
					return err
				}

				if err := htmx.IsValidGroupPassword(request.Password); err != nil {
					return err
				}

				if err := htmx.IsValidGroupDescription(request.Description); err != nil {
					return err
				}

				if err := htmx.IsValidGroupMode(request.Mode); err != nil {
					return err
				}

				group.Moniker = request.Nick
				group.Name = request.Name
				group.Description = request.Description
				group.Mode = model.GroupMode(request.Mode)
				group.Password = request.Password
				if !group.Update() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
			perms.UseForm(&GroupChange{}),
		).
		Post("/:group_id/messages/create",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[*MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				message := model.NewMessageFilled(db, group.Id, user.Id, request.Content)
				if err := htmx.IsValidMessageContent(message.Content); err != nil {
					return err
				}

				if !message.Insert() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", &fiber.Map{
						"MessageList": []model.Message{message},
						"Group":       group,
						"User":        user,
					})
					if err != nil {
						return err
					}
					str := buf.String()

					session.UserSessionMap.Push(
						render.WrapOob("beforeend:#chat", &str),
						session.PickMessages,
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
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanWriteMessages()
			}),
			perms.UseForm(&MessageCreate{}),
		).
		Post("/create",
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[*GroupCreate](ctx, perms.LocalForm)
				groupFound, _ := model.NewGroupFromName(db, request.Name)
				if !groupFound {
					return htmx.ErrGroupNotFound
				}

				if err := htmx.IsValidGroupName(request.Name); err != nil {
					return err
				}

				if err := htmx.IsValidGroupNick(request.Nick); err != nil {
					return err
				}

				if err := htmx.IsValidGroupPassword(request.Password); err != nil {
					return err
				}

				if err := htmx.IsValidGroupDescription(request.Description); err != nil {
					return err
				}

				if err := htmx.IsValidGroupMode(request.Mode); err != nil {
					return err
				}

				// TODO: validate group avatar

				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := model.NewGroup(db, user.Id, request.Nick, request.Name, model.GroupMode(request.Mode), request.Password, request.Description, request.Avatar)
				if !group.Insert() {
					return htmx.ErrDatabase
				}

				ctx.Locals(perms.LocalGroup, group)

				member := model.NewMember(db, group.Id, user.Id, "")
				if !member.Insert() {
					return htmx.ErrDatabase
				}

				everyone := model.NewRoleEveryone(db, group.Id)
				if !everyone.Insert() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, group.Url(ctx))
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.UserByAuth(db),
			perms.UseForm(&GroupCreate{}),
		).
		Get("/:group_id/messages/page/:messages_page",
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
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
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
		).
		Get("/:group_id/members/page/:members_page",
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
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
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanDeleteMessages()
			}),
		).
		Delete("/:group_id/leave",
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)

				isAlone := group.MembersCount() == 1
				if group.OwnerId == member.UserId && !isAlone {
					return htmx.ErrGroupMemberIsOnlyOwner
				}

				if isAlone {
					group.IsDeleted = true
					group.Update()
				}

				member.IsDeleted = true
				if !member.Update() {
					return htmx.ErrDatabase
				}
				if !member.LeaveActed() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, "/chat")
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return true
			}),
		).
		Delete("/:group_id",
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				group.IsDeleted = true
				if !group.Update() {
					return htmx.ErrDatabase
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
		).
		Get("/:group_id",
			func(ctx fiber.Ctx) error {
				if !websocket.IsWebSocketUpgrade(ctx) {
					return ctx.Next()
				}

				ctxWs := session.New(ctx)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				return websocket.New(func(conn *websocket.Conn) {
					ctxWs.Conn = conn
					session.UserSessionMap.Connect(user.Id, ctxWs)
					defer session.UserSessionMap.Close(user.Id, ctxWs)
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
			perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
		)
}
