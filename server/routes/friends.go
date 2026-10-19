package routes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/gofiber/fiber/v3"
)

type SendFriendRequestForm struct {
	Username     string `form:"username"`
	TargetUserId uint64 `form:"target_user_id"`
}

func RouteFriends(app *fiber.App) {
	app.Group("/friends", perms.UserByAuth()).
		Get("/list", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			return renderFriendsList(ctx, caller)
		}).
		Post("/request", perms.UseBind[SendFriendRequestForm](), func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			form := fiber.Locals[SendFriendRequestForm](ctx, perms.LocalForm)

			var targetUser model.User
			var err error

			if form.TargetUserId != 0 {
				targetUser, err = model.NewUserFromId(form.TargetUserId)
			} else {
				uInput := strings.TrimSpace(form.Username)
				uInput = strings.TrimPrefix(uInput, "@")
				if uInput == "" {
					return htmx.TryRenderPage(ctx, "partials/alert", fiber.Map{
						"Variant": "warning",
						"Message": "Please enter a valid username",
					})
				}

				// Try by handle/name
				targetUser, err = model.NewUserFromName(model.Name(uInput))
				if err != nil || targetUser.IsEmpty() {
					// Try by ID if numeric
					if idNum, errNum := strconv.ParseUint(uInput, 10, 64); errNum == nil && idNum > 0 {
						targetUser, err = model.NewUserFromId(idNum)
					}
				}
			}

			if err != nil || targetUser.IsEmpty() {
				if htmx.IsHtmx(ctx) {
					return ctx.Render("partials/alert", fiber.Map{
						"Variant": "warning",
						"Message": "User not found",
					})
				}
				return htmx.AlertUserNotFound
			}

			if targetUser.Id == caller.Id {
				if htmx.IsHtmx(ctx) {
					return ctx.Render("partials/alert", fiber.Map{
						"Variant": "warning",
						"Message": "You cannot send a friend request to yourself",
					})
				}
				return htmx.AlertFormatName
			}

			_, errReq := model.SendFriendRequest(caller.Id, targetUser.Id)
			if errReq != nil {
				if htmx.IsHtmx(ctx) {
					return ctx.Render("partials/alert", fiber.Map{
						"Variant": "warning",
						"Message": errReq.Error(),
					})
				}
				return htmx.AlertFormatName
			}

			notif := model.NewNotification(
				targetUser.Id,
				caller.Id,
				0, 0, 0,
				"friend_request",
				"Friend Request",
				fmt.Sprintf("%s (@%s) sent you a friend request.", caller.Moniker, caller.Name),
			)
			_ = model.CreateAndPushNotification(app, notif)

			if htmx.IsHtmx(ctx) {
				alertHtml, _ := render.RenderBuffer(app, "partials/alert", fiber.Map{
					"Variant": "success",
					"Message": fmt.Sprintf("Friend request sent to %s!", targetUser.Moniker),
				})
				alertStr := alertHtml.String()
				friendsOob := renderFriendsListOob(app, ctx, caller)

				// If from user profile modal
				if ctx.Get("HX-Target") == "user-profile-modal-content" {
					return renderUserProfileModal(ctx, targetUser.Id)
				}

				return ctx.Type("html").SendString(alertStr + friendsOob)
			}

			return ctx.SendStatus(fiber.StatusOK)
		}).
		Post("/requests/:id/accept", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			notifId := fiber.Params[uint64](ctx, "id")

			notifs, _ := model.GetUserNotifications(caller.Id, 100)
			var targetNotif *model.Notification
			for _, n := range notifs {
				if n.Id == notifId {
					targetNotif = &n
					break
				}
			}

			if targetNotif != nil && targetNotif.Type == "friend_request" {
				_ = model.AcceptFriendRequest(caller.Id, targetNotif.ActorId)
				_ = model.DeleteNotification(notifId, caller.Id)

				requester, errU := model.NewUserFromId(targetNotif.ActorId)
				if errU == nil && !requester.IsEmpty() {
					acceptNotif := model.NewNotification(
						requester.Id,
						caller.Id,
						0, 0, 0,
						"friend_request_accepted",
						"Friend Request Accepted",
						fmt.Sprintf("%s (@%s) accepted your friend request.", caller.Moniker, caller.Name),
					)
					_ = model.CreateAndPushNotification(app, acceptNotif)
				}
			}

			notifStateResponse := renderNotificationStateResponseStr(app, ctx, caller)
			friendsOob := renderFriendsListOob(app, ctx, caller)
			return ctx.Type("html").SendString(notifStateResponse + friendsOob)
		}).
		Post("/requests/:id/decline", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			notifId := fiber.Params[uint64](ctx, "id")

			notifs, _ := model.GetUserNotifications(caller.Id, 100)
			var targetNotif *model.Notification
			for _, n := range notifs {
				if n.Id == notifId {
					targetNotif = &n
					break
				}
			}

			if targetNotif != nil && targetNotif.Type == "friend_request" {
				_ = model.DeclineOrCancelFriendRequest(caller.Id, targetNotif.ActorId)
				_ = model.DeleteNotification(notifId, caller.Id)
			}

			notifStateResponse := renderNotificationStateResponseStr(app, ctx, caller)
			friendsOob := renderFriendsListOob(app, ctx, caller)
			return ctx.Type("html").SendString(notifStateResponse + friendsOob)
		}).
		Post("/accept/:target_id", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			targetId := fiber.Params[uint64](ctx, "target_id")

			_ = model.AcceptFriendRequest(caller.Id, targetId)

			// Clean up any friend request notifications from this target user
			notifs, _ := model.GetUserNotifications(caller.Id, 100)
			for _, n := range notifs {
				if n.ActorId == targetId && n.Type == "friend_request" {
					_ = model.DeleteNotification(n.Id, caller.Id)
				}
			}

			targetUser, errU := model.NewUserFromId(targetId)
			if errU == nil && !targetUser.IsEmpty() {
				acceptNotif := model.NewNotification(
					targetUser.Id,
					caller.Id,
					0, 0, 0,
					"friend_request_accepted",
					"Friend Request Accepted",
					fmt.Sprintf("%s (@%s) accepted your friend request.", caller.Moniker, caller.Name),
				)
				_ = model.CreateAndPushNotification(app, acceptNotif)
			}

			if ctx.Get("HX-Target") == "user-profile-modal-content" {
				return renderUserProfileModal(ctx, targetId)
			}

			friendsOob := renderFriendsListOob(app, ctx, caller)
			return ctx.Type("html").SendString(friendsOob)
		}).
		Post("/decline/:target_id", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			targetId := fiber.Params[uint64](ctx, "target_id")

			_ = model.DeclineOrCancelFriendRequest(caller.Id, targetId)

			// Clean up any friend request notifications from this target user
			notifs, _ := model.GetUserNotifications(caller.Id, 100)
			for _, n := range notifs {
				if n.ActorId == targetId && n.Type == "friend_request" {
					_ = model.DeleteNotification(n.Id, caller.Id)
				}
			}

			if ctx.Get("HX-Target") == "user-profile-modal-content" {
				return renderUserProfileModal(ctx, targetId)
			}

			friendsOob := renderFriendsListOob(app, ctx, caller)
			return ctx.Type("html").SendString(friendsOob)
		}).
		Delete("/:target_id", func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			targetId := fiber.Params[uint64](ctx, "target_id")

			_ = model.RemoveFriend(caller.Id, targetId)

			if ctx.Get("HX-Target") == "user-profile-modal-content" {
				return renderUserProfileModal(ctx, targetId)
			}

			return renderFriendsList(ctx, caller)
		})
}

func renderFriendsList(ctx fiber.Ctx, user model.User) error {
	bind := MapPage(ctx, fiber.Map{
		"User": user,
	})
	return ctx.Render("partials/friends-list", bind)
}

func renderFriendsListOob(app *fiber.App, ctx fiber.Ctx, user model.User) string {
	buf, err := render.RenderBuffer(app, "partials/friends-list", MapPage(ctx, fiber.Map{
		"User": user,
	}))
	if err != nil {
		return ""
	}
	s := buf.String()
	return render.WrapOob("innerHTML:#friends-list", &s)
}

func renderNotificationStateResponseStr(app *fiber.App, ctx fiber.Ctx, user model.User) string {
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
		return ""
	}
	s := buf.String()
	dropdownOob := render.WrapOob("innerHTML:#dropdown-notification-list", &s)
	pageOob := render.WrapOob("innerHTML:#page-notification-list", &s)

	return badgeOob + dropdownOob + pageOob
}
