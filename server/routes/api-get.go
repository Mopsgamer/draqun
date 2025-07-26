package routes

import (
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func RegisterGetRoutes(app *fiber.App, db *database.DB) {
	app.Get("/groups/:group_id/messages/page/:messages_page",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
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
		perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.PermMessages.CanReadMessages()
		}),
	)
	app.Get("/groups/:group_id/members/page/:members_page",
		func(ctx fiber.Ctx) error {
			group := fiber.Locals[database.Group](ctx, perms.LocalGroup)
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
		perms.MemberByAuthAndGroupId(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.PermMessages.CanDeleteMessages()
		}),
	)
}
