package routes

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/gofiber/fiber/v3"
)

func RouteNotifications(app *fiber.App) fiber.Router {
	return app.Group("/notifications", perms.UserByAuth()).
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

	return ctx.Type("html").SendString(badgeOob + dropdownOob + pageOob)
}
