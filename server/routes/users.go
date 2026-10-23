package routes

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

type SuspendForm struct {
	Reason  string `form:"reason"`
	GroupId uint64 `form:"group_id"`
}

type BlockIPForm struct {
	IP      string `form:"ip"`
	Reason  string `form:"reason"`
	GroupId uint64 `form:"group_id"`
}

func renderUserProfileModal(ctx fiber.Ctx, targetUserId uint64) error {
	caller := fiber.Locals[model.User](ctx, perms.LocalAuth)

	targetUser, err := model.NewUserFromId(targetUserId)
	if err != nil || targetUser.IsEmpty() {
		return htmx.AlertUserNotFound
	}

	groupIdStr := ctx.Query("group_id", ctx.FormValue("group_id", ""))
	var group model.Group
	var member model.Member
	var roles []model.Role
	var groupJoinDateStr string
	var canKickBanGroup bool
	var groupMessageCount int

	if groupIdStr != "" {
		groupId, _ := strconv.ParseUint(groupIdStr, 10, 64)
		if groupId != 0 {
			g, errG := model.NewGroupFromId(groupId)
			if errG == nil && !g.IsEmpty() {
				group = g
				m, errM := model.NewMemberFromId(group.Id, targetUser.Id)
				if errM == nil && m.IsAvailable() {
					member = m
					roles = m.Roles()
					groupJoinDateStr = time.Time(m.FirstTimeJoinedAt).Format("Jan 02, 2006 15:04")
					groupMessageCount = int(m.MessagesCount())
				}

				callerLevel := model.GetUserPermissionLevel(group.Id, caller.Id, 0, model.PermNameMembers)
				if callerLevel != model.PermLevelDisable || caller.Id == group.OwnerId {
					canKickBanGroup = true
				}
			}
		}
	}

	isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)

	var targetIPs []string
	blockedIPsMap := make(map[string]bool)
	var targetLogs []model.UserLog

	if isDraqunAdmin || isDraqunMod {
		targetIPs = model.GetUserIPs(targetUser.Id)
		for _, ip := range targetIPs {
			if model.IsIPBlocked(ip) {
				blockedIPsMap[ip] = true
			}
		}
		targetLogs = model.GetUserLogs(targetUser.Id, 20)
	}

	accountCreatedStr := time.Time(targetUser.CreatedAt).Format("Jan 02, 2006 15:04")
	lastOnlineStr := time.Time(targetUser.LastSeenAt).Format("Jan 02, 2006 15:04")

	friendship := caller.FriendshipWith(targetUser.Id)
	var friendshipStatus string
	var isFriendRequester bool
	if friendship != nil && !friendship.IsEmpty() {
		friendshipStatus = friendship.Status
		if friendship.UserId == caller.Id {
			isFriendRequester = true
		}
	}

	bind := fiber.Map{
		"TargetUser":         targetUser,
		"Caller":             caller,
		"Group":              group,
		"TargetMember":       member,
		"TargetRoles":        roles,
		"GroupJoinDate":      groupJoinDateStr,
		"GroupMessageCount":  groupMessageCount,
		"CanKickBanGroup":    canKickBanGroup,
		"IsDraqunAdmin":      isDraqunAdmin,
		"IsDraqunMod":        isDraqunMod,
		"TargetIPs":          targetIPs,
		"BlockedIPsMap":      blockedIPsMap,
		"TargetLogs":         targetLogs,
		"AccountCreatedDate": accountCreatedStr,
		"LastOnlineDate":     lastOnlineStr,
		"FriendshipStatus":   friendshipStatus,
		"IsFriendRequester":  isFriendRequester,
	}

	if htmx.IsHtmx(ctx) {
		return ctx.Render("partials/user-profile-modal", MapPage(ctx, bind))
	}

	return ctx.JSON(bind)
}

func RouteUsers(app *fiber.App) {
	app.Get("/users/:user_id/profile",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			targetUserId := fiber.Params[uint64](ctx, "user_id")
			return renderUserProfileModal(ctx, targetUserId)
		},
	)

	app.Post("/admin/users/:user_id/reset-avatar",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			targetUser.Avatar = ""
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "reset_avatar", "Avatar reset to default value")

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/reset-moniker",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			targetUser.Moniker = model.Moniker(fmt.Sprintf("User_%d", targetUser.Id))
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "reset_moniker", fmt.Sprintf("Moniker reset to User_%d", targetUser.Id))

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/reset-name",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			targetUser.Name = model.Name(fmt.Sprintf("user_%d", targetUser.Id))
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "reset_name", fmt.Sprintf("Username (@name) reset to user_%d", targetUser.Id))

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/suspend",
		perms.UserByAuth(),
		perms.UseBind[SuspendForm](),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			form := fiber.Locals[SuspendForm](ctx, perms.LocalForm)

			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			targetUser.IsSuspended = true
			targetUser.SuspensionReason = form.Reason
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "suspend", "Account suspended: "+form.Reason)

			reasonStr := strings.TrimSpace(form.Reason)
			var notifContent string
			if reasonStr != "" {
				notifContent = fmt.Sprintf("Reason: %s (Duration: Indefinite / Permanent)", reasonStr)
			} else {
				notifContent = "Duration: Indefinite / Permanent"
			}
			notif := model.NewNotification(
				targetUser.Id,
				caller.Id,
				form.GroupId,
				0,
				0,
				"banned",
				"Your account has been suspended",
				notifContent,
			)
			_ = model.CreateAndPushNotification(app, notif)

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/unsuspend",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			targetUser.IsSuspended = false
			targetUser.SuspensionReason = ""
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "unsuspend", "Account suspension lifted")

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/lockdown",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			if !targetUser.IsLocked {
				targetUser.LockedPassword = targetUser.Password
			}
			scrambledHash, errHash := model.Password(fmt.Sprintf("locked_%d_%d", targetUser.Id, time.Now().UnixNano())).Hash()
			if errHash == nil {
				targetUser.Password = scrambledHash
			}
			targetUser.IsLocked = true

			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "lockdown", "Account locked down and password reset for security")

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/users/:user_id/unlock",
		perms.UserByAuth(),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			targetUserId := fiber.Params[uint64](ctx, "user_id")
			targetUser, err := model.NewUserFromId(targetUserId)
			if err != nil || targetUser.IsEmpty() {
				return htmx.AlertUserNotFound
			}

			if targetUser.LockedPassword != "" {
				targetUser.Password = targetUser.LockedPassword
				targetUser.LockedPassword = ""
			}
			targetUser.IsLocked = false
			if err := targetUser.Update(); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			model.RecordUserLog(targetUser.Id, caller.Id, "unlock", "Account unlocked")

			if htmx.IsHtmx(ctx) {
				return renderUserProfileModal(ctx, targetUser.Id)
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/ips/block",
		perms.UserByAuth(),
		perms.UseBind[BlockIPForm](),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			form := fiber.Locals[BlockIPForm](ctx, perms.LocalForm)

			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			if err := model.BlockIP(form.IP, caller.Id, form.Reason); err != nil {
				return htmx.AlertFormatMoniker.Join(err)
			}

			targetUserIdStr := ctx.FormValue("user_id")
			if targetUserIdStr != "" {
				targetUserId, _ := strconv.ParseUint(targetUserIdStr, 10, 64)
				if targetUserId != 0 {
					model.RecordUserLog(targetUserId, caller.Id, "block_ip", fmt.Sprintf("Blocked IP %s: %s", form.IP, form.Reason))
					if htmx.IsHtmx(ctx) {
						return renderUserProfileModal(ctx, targetUserId)
					}
				}
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)

	app.Post("/admin/ips/unblock",
		perms.UserByAuth(),
		perms.UseBind[BlockIPForm](),
		func(ctx fiber.Ctx) error {
			caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
			form := fiber.Locals[BlockIPForm](ctx, perms.LocalForm)

			isDraqunAdmin, isDraqunMod := model.IsDraqunAdminOrMod(caller)
			if !isDraqunAdmin && !isDraqunMod {
				return htmx.AlertGroupMemberNotAllowed
			}

			if err := model.UnblockIP(form.IP); err != nil {
				return htmx.AlertDatabase.Join(err)
			}

			targetUserIdStr := ctx.FormValue("user_id")
			if targetUserIdStr != "" {
				targetUserId, _ := strconv.ParseUint(targetUserIdStr, 10, 64)
				if targetUserId != 0 {
					model.RecordUserLog(targetUserId, caller.Id, "unblock_ip", fmt.Sprintf("Unblocked IP %s", form.IP))
					if htmx.IsHtmx(ctx) {
						return renderUserProfileModal(ctx, targetUserId)
					}
				}
			}

			return ctx.SendStatus(fiber.StatusOK)
		},
	)
}
