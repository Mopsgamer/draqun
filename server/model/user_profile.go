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

	mandatoryChannels := []struct {
		Name        Name
		Moniker     Moniker
		Description Description
		Type        ChannelType
	}{
		{Name("general"), Moniker("General"), Description("General discussion channel"), ChannelTypeText},
		{Name("bugs"), Moniker("Bugs"), Description("Bug reports ticket channel"), ChannelTypeTicket},
		{Name("feats"), Moniker("Feats"), Description("Feature requests ticket channel"), ChannelTypeTicket},
		{Name("reports"), Moniker("Reports"), Description("User reports ticket channel"), ChannelTypeTicket},
	}

	for i, chSpec := range mandatoryChannels {
		ch := NewChannel(chSpec.Name, chSpec.Moniker, chSpec.Description, chSpec.Type)
		if err := ch.Insert(); err == nil {
			_ = AddGroupChannel(newGroup.Id, ch.Id, uint(i+1))
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

	member, err := NewMemberFromId(draqunGroup.Id, actor.Id)
	if err == nil && member.IsAvailable() {
		role := member.Role()
		if role.PermAdmin.Has() {
			return true, true
		}
		if role.PermMembers.CanKickBan() {
			return false, true
		}
	}

	return false, false
}
