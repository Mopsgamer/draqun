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

type GroupCreate struct {
	Name        model.Name             `form:"name"`
	Nick        model.Moniker          `form:"nick"`
	Password    model.OptionalPassword `form:"password"`
	Mode        model.GroupMode        `form:"mode"`
	Description model.Description      `form:"description"`
	Avatar      model.Avatar           `form:"avatar"`
}

type MessageCreate struct {
	Content string `form:"content"`
}

type GroupChange struct {
	Name        model.Name             `form:"name"`
	Nick        model.Moniker          `form:"nick"`
	Password    model.OptionalPassword `form:"password"`
	Mode        model.GroupMode        `form:"mode"`
	Description model.Description      `form:"description"`
	Avatar      model.Avatar           `form:"avatar"`
}

func RouteGroups(app *fiber.App) fiber.Router {
	return app.Group("/groups").
		Put("/:group_id/join",
			perms.GroupById("group_id"),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)
				if member.IsEmpty() {
					// never been a member
					member = model.NewMember(group.Id, member.UserId, "")
					if err := member.Insert(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				} else if member.IsDeleted {
					// been a member, now isn't
					member.IsDeleted = false
					if err := member.Update(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				} else {
					// already a member
					return htmx.AlertUseless
				}

				if err := member.JoinActed(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					// controller_ws.UserSessionMap.Push(
					// 	filter,
					// 	controller.WrapOob("beforeend:#chat", &str),
					// 	controller_ws.SubForMessages,
					// )

					htmx.Redirect(ctx, group.Url(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/:group_id/change",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
			perms.UseBind[GroupChange](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[GroupChange](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				hasChanges := group.Moniker != request.Nick ||
					group.Name != request.Name ||
					group.Description != request.Description ||
					group.Mode != request.Mode ||
					group.Password.Compare(request.Password) != nil

				if !hasChanges {
					return htmx.AlertUseless
				}

				group.Moniker = request.Nick
				group.Name = request.Name
				group.Description = request.Description
				group.Mode = model.GroupMode(request.Mode)
				var err error
				group.Password, err = request.Password.Hash()
				if err != nil {
					return htmx.AlertEncryption.Join(err)
				}

				if err := group.Validate(); err != nil {
					return err
				}

				if err := group.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/messages/create",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanWriteMessages()
			}),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				message := model.NewMessageFilled(group.Id, user.Id, request.Content)
				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()

					session.UserSessionMap.Push(
						render.WrapOob("beforeend:#chat", &str),
						session.PickMessages,
					)

					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/create",
			perms.UserByAuth(),
			perms.UseBind[GroupCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[GroupCreate](ctx, perms.LocalForm)
				existingGroup, _ := model.NewGroupFromName(request.Name)
				if !existingGroup.IsEmpty() {
					return htmx.AlertGroupExistsName
				}

				hash, err := request.Password.Hash()
				if err != nil {
					return htmx.AlertEncryption.Join(err)
				}

				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := model.NewGroup(user.Id, request.Nick, request.Name, request.Mode, hash, request.Description, request.Avatar)
				if err := group.Validate(); err != nil {
					return err
				}

				if err := group.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				ctx.Locals(perms.LocalGroup, group)

				member := model.NewMember(group.Id, user.Id, "")
				if err := member.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				everyone := model.NewRoleEveryone(group.Id)
				if err := everyone.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, group.Url(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/:group_id/messages/page/:messages_page",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
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
		).
		Get("/:group_id/members/page/:members_page",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMembers.CanSee()
			}),
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
		).
		Delete("/:group_id/leave",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return true
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)

				isAlone := group.MembersCount() == 1
				if group.OwnerId == member.UserId && !isAlone {
					return htmx.AlertGroupMemberIsOnlyOwner
				}

				if isAlone {
					group.IsDeleted = true
					if err := group.Update(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				}

				member.IsDeleted = true
				if err := member.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}
				if err := member.LeaveActed(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, "/chat")
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				group.IsDeleted = true
				if err := group.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/ws/:group_id/",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
			func(ctx fiber.Ctx) error {
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
					_ = ctxWs.Conn.Close()
				})(ctx)
			},
		)
}
