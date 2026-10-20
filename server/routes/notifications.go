package routes

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func RouteNotifications(app *fiber.App) fiber.Router {
	return app.Group("/notifications", perms.UserByAuth()).
		Get("/ws", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			ctxWs := session.New(ctx)

			return websocket.New(func(conn *websocket.Conn) {
				ctxWs.Conn = conn
				session.UserSessionMap.Connect(user.Id, ctxWs)
				BroadcastUserPresenceUpdate(app, user.Id)
				defer func() {
					session.UserSessionMap.Close(user.Id, ctxWs)
					BroadcastUserPresenceUpdate(app, user.Id)
					ctxWs.Closed = true
					_ = ctxWs.Conn.Close()
				}()
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
			})(ctx)
		}).
		Get("", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			notifications, err := model.GetUserNotifications(user.Id, 50)
			if err != nil {
				notifications = []model.Notification{}
			}
			unreadCount, _ := model.GetUnreadNotificationCount(user.Id)

			bind := fiber.Map{
				"Title":         "Notifications",
				"Notifications": notifications,
				"UnreadCount":   unreadCount,
				"User":          user,
			}

			if htmx.IsHtmx(ctx) {
				return ctx.Render("partials/notifications", bind)
			}

			return htmx.TryRenderPage(ctx, "notifications", MapPage(ctx, bind), "partials/main")
		}).
		Post("/:id/read", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			notifId := fiber.Params[uint64](ctx, "id")

			_ = model.MarkNotificationRead(notifId, user.Id)
			return renderNotificationStateResponse(app, ctx, user)
		}).
		Post("/read-all", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)

			_ = model.MarkAllNotificationsRead(user.Id)
			return renderNotificationStateResponse(app, ctx, user)
		}).
		Delete("/all", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)

			_ = model.DeleteAllNotifications(user.Id)
			if htmx.IsHtmx(ctx) && ctx.Get("HX-Target") == "body" {
				htmx.Redirect(ctx, "/notifications")
				return nil
			}
			return renderNotificationStateResponse(app, ctx, user)
		}).
		Delete("/:id", func(ctx fiber.Ctx) error {
			user := fiber.Locals[model.User](ctx, perms.LocalAuth)
			notifId := fiber.Params[uint64](ctx, "id")

			_ = model.DeleteNotification(notifId, user.Id)
			return renderNotificationStateResponse(app, ctx, user)
		})
}

func renderNotificationStateResponse(app *fiber.App, ctx fiber.Ctx, user model.User) error {
	unreadCount, _ := model.GetUnreadNotificationCount(user.Id)
	notifications, _ := model.GetUserNotifications(user.Id, 50)

	bind := fiber.Map{
		"Notifications": notifications,
		"UnreadCount":   unreadCount,
		"User":          user,
	}

	badgeOob := render.RenderUnreadBadgeOob(unreadCount)

	buf, err := render.RenderBuffer(app, "partials/notifications", bind)
	if err != nil {
		return err
	}
	s := buf.String()
	dropdownOob := render.WrapOob("innerHTML:#dropdown-notification-list", &s)
	pageOob := render.WrapOob("innerHTML:#page-notification-list", &s)

	return ctx.Type("html").SendString(s + badgeOob + dropdownOob + pageOob)
}
