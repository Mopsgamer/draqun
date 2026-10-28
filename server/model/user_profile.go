package model

import (
	"errors"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type UserLog struct {
	checkEmpty
	Id        uint64   `db:"id" goqu:"skipinsert"`
	UserId    uint64   `db:"user_id"`
	ActorId   uint64   `db:"actor_id"`
	Action    string   `db:"action"`
	Details   string   `db:"details"`
	CreatedAt TimePast `db:"created_at"`
	Actor     User     `db:"-"`
}

func RecordUserIP(userId uint64, ip string) {
	ip = strings.TrimSpace(ip)
	if userId == 0 || ip == "" {
		return
	}

	// Update last_ip on app_users
	_, _ = Sqlx.Exec("UPDATE app_users SET last_ip = ? WHERE id = ?", ip, userId)

	// Insert or replace into app_user_ips
	now := time.Now().Format("2006-01-02 15:04:05")
	_, _ = Sqlx.Exec("INSERT INTO app_user_ips (user_id, ip, last_seen_at) VALUES (?, ?, ?) ON CONFLICT(user_id, ip) DO UPDATE SET last_seen_at = ?", userId, ip, now, now)
}

func GetUserIPs(userId uint64) []string {
	var ips []string
	if userId == 0 {
		return ips
	}
	sql, args, err := Goqu.From(TableUserIPs).Select("ip").Where(goqu.Ex{"user_id": userId}).Order(goqu.I("last_seen_at").Desc()).Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return ips
	}
	_ = Sqlx.Select(&ips, sql, args...)
	return ips
}

func BlockIP(ip string, creatorId uint64, reason string) error {
	ip = strings.TrimSpace(ip)
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return errors.New("cannot block localhost IP address")
	}
	if ip == "" {
		return errors.New("invalid IP address")
	}

	_, err := Sqlx.Exec("INSERT INTO app_blocked_ips (ip, creator_id, reason, created_at) VALUES (?, ?, ?, ?) ON CONFLICT(ip) DO UPDATE SET reason = ?", ip, creatorId, reason, time.Now().Format("2006-01-02 15:04:05"), reason)
	return err
}

func UnblockIP(ip string) error {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil
	}
	_, err := Sqlx.Exec("DELETE FROM app_blocked_ips WHERE ip = ?", ip)
	return err
}

func IsIPBlocked(ip string) bool {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return false
	}
	var count int
	err := Sqlx.Get(&count, "SELECT COUNT(*) FROM app_blocked_ips WHERE ip = ?", ip)
	if err != nil {
		return false
	}
	return count > 0
}

func RecordUserLog(userId uint64, actorId uint64, action string, details string) {
	if userId == 0 {
		return
	}
	now := TimePast(time.Now())
	logEntry := UserLog{
		UserId:    userId,
		ActorId:   actorId,
		Action:    action,
		Details:   details,
		CreatedAt: now,
	}
	_ = Insert(TableUserLogs, &logEntry)
}

func GetUserLogs(userId uint64, limit int) []UserLog {
	var logs []UserLog
	if userId == 0 {
		return logs
	}
	if limit <= 0 {
		limit = 20
	}
	sql, args, err := Goqu.From(TableUserLogs).Where(goqu.Ex{"user_id": userId}).Order(goqu.I("id").Desc()).Limit(uint(limit)).Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return logs
	}
	_ = Sqlx.Select(&logs, sql, args...)

	for i := range logs {
		if logs[i].ActorId != 0 {
			actor, _ := NewUserFromId(logs[i].ActorId)
			logs[i].Actor = actor
		}
	}
	return logs
}

func GetOrCreateDraqunGroup(ownerId uint64) (Group, error) {
	group, err := NewGroupFromName("draqun")
	if err == nil && !group.IsEmpty() {
		return group, nil
	}

	if ownerId == 0 {
		ownerId = 1
	}

	newGroup := NewGroup(
		ownerId,
		Moniker("Draqun"),
		Name("draqun"),
		GroupModePublic,
		OptionalPasswordHashed(""),
		Description("Official Draqun Community Group"),
		Avatar(""),
	)
	if err := newGroup.Insert(); err != nil {
		return newGroup, err
	}

	member := NewMember(newGroup.Id, ownerId, "")
	_ = member.Insert()

	everyone := NewRoleEveryone(newGroup.Id)
	_ = everyone.Insert()

	adminRole := Role{
		GroupId:  newGroup.Id,
		Position: 1,
		Name:     "admin",
		Moniker:  "Admin",
		Color:    ParseColorHex("#ef4444"),
	}
	_ = adminRole.Insert()

	supportRole := Role{
		GroupId:  newGroup.Id,
		Position: 2,
		Name:     "support",
		Moniker:  "Support",
		Color:    ParseColorHex("#3b82f6"),
	}
	_ = supportRole.Insert()

	punisherRole := Role{
		GroupId:  newGroup.Id,
		Position: 3,
		Name:     "punisher",
		Moniker:  "Punisher",
		Color:    ParseColorHex("#a855f7"),
	}
	_ = punisherRole.Insert()

	if ownerId != 0 && adminRole.Id != 0 {
		ra := RoleAssignee{
			RoleId: adminRole.Id,
			UserId: ownerId,
		}
		_ = ra.Insert()
	}

	if adminRole.Id != 0 {
		adminGroupPerm := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameAdmin,
			Level:    PermLevelAll,
			Priority: 10,
		}
		if err := adminGroupPerm.Insert(); err == nil {
			_ = adminGroupPerm.AddRole(adminRole.Id)
		}
	}

	mandatoryChannels := []struct {
		Name        Name
		Moniker     Moniker
		Description Description
		Type        ChannelType
	}{
		{Name("general"), Moniker("General"), Description("The main hub for community conversations, casual chat, and general discussions."), ChannelTypeText},
		{Name("internal"), Moniker("Internal"), Description("Private internal discussion channel for moderators and staff members."), ChannelTypeText},
		{Name("bugs"), Moniker("Bugs"), Description("Report bugs, technical glitches, and platform issues to help us fix them."), ChannelTypeTicket},
		{Name("feats"), Moniker("Feats"), Description("Propose new features, improvements, and ideas to shape the future of Draqun."), ChannelTypeTicket},
		{Name("admin-requests"), Moniker("Admin Requests"), Description("Internal channel for moderators and staff to submit escalation requests directly to administrators."), ChannelTypeTicket},
		{Name("private-support"), Moniker("Private Support"), Description("Open a private ticket to get 1-on-1 assistance from the support team for account or personal issues."), ChannelTypeTicket},
		{Name("public-support"), Moniker("Public Support"), Description("Community-wide support channel for common questions and shared knowledge, managed by support staff."), ChannelTypeTicket},
		{Name("otel"), Moniker("OpenTelemetry"), Description("OpenTelemetry traces, metrics, and server performance stats for Draqun system administrators."), ChannelTypeOtel},
	}

	createdChannels := make(map[Name]Channel)
	for i, chSpec := range mandatoryChannels {
		ch := NewChannel(chSpec.Name, chSpec.Moniker, chSpec.Description, chSpec.Type)
		if err := ch.Insert(); err == nil {
			_ = AddGroupChannel(newGroup.Id, ch.Id, uint(i+1))
			createdChannels[chSpec.Name] = ch
		}
	}

	// 0) Internal text channel: only support and punisher have read/write access, isAdmin has full access, everyone has no access
	if internalCh, ok := createdChannels["internal"]; ok {
		for _, permName := range []PermName{PermNameRead, PermNameWrite, PermNameDelete} {
			pNoEveryone := Permission{
				GroupId:  newGroup.Id,
				Name:     permName,
				Level:    PermLevelDisable,
				Priority: 10,
			}
			if err := pNoEveryone.Insert(); err == nil {
				_ = pNoEveryone.AddChannel(internalCh.Id)
				_ = pNoEveryone.AddRole(everyone.Id)
			}

			pReadAllStaff := Permission{
				GroupId:  newGroup.Id,
				Name:     permName,
				Level:    PermLevelAll,
				Priority: 20,
			}
			if err := pReadAllStaff.Insert(); err == nil {
				_ = pReadAllStaff.AddChannel(internalCh.Id)
				if supportRole.Id != 0 {
					_ = pReadAllStaff.AddRole(supportRole.Id)
				}
				if punisherRole.Id != 0 {
					_ = pReadAllStaff.AddRole(punisherRole.Id)
				}
			}
		}
	}

	// 1) Admin requests: only support and punisher have read/write access, isAdmin has full access, everyone has no access
	if adminReqCh, ok := createdChannels["admin-requests"]; ok {
		for _, permName := range []PermName{PermNameRead, PermNameWrite, PermNameDelete} {
			pNoEveryone := Permission{
				GroupId:  newGroup.Id,
				Name:     permName,
				Level:    PermLevelDisable,
				Priority: 10,
			}
			if err := pNoEveryone.Insert(); err == nil {
				_ = pNoEveryone.AddChannel(adminReqCh.Id)
				_ = pNoEveryone.AddRole(everyone.Id)
			}

			pReadAllStaff := Permission{
				GroupId:  newGroup.Id,
				Name:     permName,
				Level:    PermLevelAll,
				Priority: 20,
			}
			if err := pReadAllStaff.Insert(); err == nil {
				_ = pReadAllStaff.AddChannel(adminReqCh.Id)
				if supportRole.Id != 0 {
					_ = pReadAllStaff.AddRole(supportRole.Id)
				}
				if punisherRole.Id != 0 {
					_ = pReadAllStaff.AddRole(punisherRole.Id)
				}
			}
		}
	}

	// 2) Private Support: everyone read own, support staff read all
	if privSuppCh, ok := createdChannels["private-support"]; ok {
		pReadOwn := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameRead,
			Level:    PermLevelOwn,
			Priority: 10,
		}
		if err := pReadOwn.Insert(); err == nil {
			_ = pReadOwn.AddChannel(privSuppCh.Id)
			_ = pReadOwn.AddRole(everyone.Id)
		}

		pReadAllStaff := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameRead,
			Level:    PermLevelAll,
			Priority: 20,
		}
		if err := pReadAllStaff.Insert(); err == nil {
			_ = pReadAllStaff.AddChannel(privSuppCh.Id)
			if supportRole.Id != 0 {
				_ = pReadAllStaff.AddRole(supportRole.Id)
			}
		}
	}

	// 3) Public Support: read all, write disabled, delete disabled. Only Support and Admin can have write access.
	if pubSuppCh, ok := createdChannels["public-support"]; ok {
		pReadAll := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameRead,
			Level:    PermLevelAll,
			Priority: 10,
		}
		if err := pReadAll.Insert(); err == nil {
			_ = pReadAll.AddChannel(pubSuppCh.Id)
			_ = pReadAll.AddRole(everyone.Id)
		}

		pWriteDisable := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameWrite,
			Level:    PermLevelDisable,
			Priority: 10,
		}
		if err := pWriteDisable.Insert(); err == nil {
			_ = pWriteDisable.AddChannel(pubSuppCh.Id)
			_ = pWriteDisable.AddRole(everyone.Id)
		}

		pDeleteDisable := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameDelete,
			Level:    PermLevelDisable,
			Priority: 10,
		}
		if err := pDeleteDisable.Insert(); err == nil {
			_ = pDeleteDisable.AddChannel(pubSuppCh.Id)
			_ = pDeleteDisable.AddRole(everyone.Id)
		}

		pWriteStaff := Permission{
			GroupId:  newGroup.Id,
			Name:     PermNameWrite,
			Level:    PermLevelAll,
			Priority: 20,
		}
		if err := pWriteStaff.Insert(); err == nil {
			_ = pWriteStaff.AddChannel(pubSuppCh.Id)
			if supportRole.Id != 0 {
				_ = pWriteStaff.AddRole(supportRole.Id)
			}
			if adminRole.Id != 0 {
				_ = pWriteStaff.AddRole(adminRole.Id)
			}
		}
	}

	// 4) OTel channel: disabled for everyone by default (route checks enforce global admin)
	if otelCh, ok := createdChannels["otel"]; ok {
		for _, permName := range []PermName{PermNameRead, PermNameWrite, PermNameDelete} {
			pNoEveryone := Permission{
				GroupId:  newGroup.Id,
				Name:     permName,
				Level:    PermLevelDisable,
				Priority: 10,
			}
			if err := pNoEveryone.Insert(); err == nil {
				_ = pNoEveryone.AddChannel(otelCh.Id)
				_ = pNoEveryone.AddRole(everyone.Id)
			}
		}
	}

	return newGroup, nil
}

func IsDraqunAdminOrMod(actor User) (isAdmin bool, isMod bool) {
	if actor.IsEmpty() {
		return false, false
	}

	if actor.Id == 1 {
		return true, true
	}

	draqunGroup, err := NewGroupFromName("draqun")
	if err != nil || draqunGroup.IsEmpty() {
		// Fallback if no draqun group exists yet: user ID 1 or group owners
		if actor.Id == 1 || len(actor.GroupListOwner()) > 0 {
			return true, true
		}
		return false, false
	}

	if actor.Id == draqunGroup.OwnerId || actor.Id == draqunGroup.CreatorId {
		return true, true
	}

	if GetUserPermissionLevel(draqunGroup.Id, actor.Id, 0, PermNameAdmin) == PermLevelAll {
		return true, true
	}
	if GetUserPermissionLevel(draqunGroup.Id, actor.Id, 0, PermNameMembers) == PermLevelAll {
		return false, true
	}

	return false, false
}
