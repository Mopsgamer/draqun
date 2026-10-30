package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

type GroupMode string

const (
	GroupModePrivate GroupMode = "private"
	GroupModePublic  GroupMode = "public"
)

func (gm GroupMode) IsValid() bool {
	return gm == GroupModePrivate || gm == GroupModePublic
}

type Group struct {
	checkEmpty
	Id          uint64                 `db:"id" goqu:"skipinsert"`
	CreatorId   uint64                 `db:"creator_id"`
	OwnerId     uint64                 `db:"owner_id"`
	Moniker     Moniker                `db:"moniker"` // Nick is a customizable name.
	Name        Name                   `db:"name"`    // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode              `db:"mode"`
	Password    OptionalPasswordHashed `db:"password"`
	Description Description            `db:"description"`
	Avatar      Avatar                 `db:"avatar"`
	CreatedAt   TimePast               `db:"created_at"`
	IsDeleted   bool                   `db:"is_deleted"`
}

var _ Model = (*Group)(nil)

func NewGroup(
	creatorId uint64,
	moniker Moniker,
	name Name,
	mode GroupMode,
	password OptionalPasswordHashed,
	description Description,
	avatar Avatar,
) Group {
	return Group{
		CreatorId:   creatorId,
		OwnerId:     creatorId,
		Moniker:     moniker,
		Name:        name,
		Mode:        mode,
		Password:    password,
		Description: description,
		Avatar:      avatar,
		CreatedAt:   TimePast(time.Now()),
		IsDeleted:   false,
	}
}

func NewGroupFromId(id uint64) (Group, error) {
	group := Group{}
	return group, First(TableGroups, goqu.Ex{"id": id}, &group)
}

func NewGroupFromName(name Name) (Group, error) {
	group := Group{}
	return group, First(TableGroups, goqu.Ex{"name": name}, &group)
}

func (group Group) Validate() htmx.Alert {
	if !group.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !group.Name.IsValid() {
		return htmx.AlertFormatName
	}
	if !group.Mode.IsValid() {
		return htmx.AlertFormatGroupMode
	}
	if !group.Password.IsValid() {
		return htmx.AlertFormatPassword
	}
	if !group.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !group.Avatar.IsValid() {
		return htmx.AlertFormatAvatar
	}
	if !group.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (group Group) IsAvailable() bool {
	return !group.IsEmpty() && !bool(group.IsDeleted)
}

func (group *Group) Insert() error {
	return InsertId(TableGroups, group, &group.Id)
}

func (group Group) Update() error {
	return Update(TableGroups, group, goqu.Ex{"id": group.Id})
}

func (group Group) Creator() User {
	user, _ := NewUserFromId(group.CreatorId)
	return user
}

func (group Group) Owner() User {
	user, _ := NewUserFromId(group.OwnerId)
	return user
}

func (group Group) Everyone() Role {
	role, _ := NewRoleFromName(roleNameEveryone, group.Id)
	return role
}

func (group Group) Roles() []Role {
	roleList := []Role{}
	sql, args, err := Goqu.From(TableRoles).
		Select("id", "group_id", "position", "name", "moniker", "color").
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("position").Asc(), goqu.I("id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return roleList
	}

	err = Sqlx.Select(&roleList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return roleList
}

func (group Group) IsBanned(userId uint64) bool {
	if userId == 0 {
		return false
	}

	user, err := NewUserFromId(userId)
	if err == nil && !user.IsEmpty() && user.IsSuspended {
		return true
	}

	nowStr := time.Now().Format(time.RFC3339)

	draqunGroup, err := NewGroupFromName("draqun")
	if err == nil && !draqunGroup.IsEmpty() && draqunGroup.Id != group.Id {
		var draqunCount uint64
		draqunSql, draqunArgs, _ := Goqu.From(TableBans).
			Select(goqu.COUNT("*")).
			Where(goqu.Ex{
				"group_id":   draqunGroup.Id,
				"target_id":  userId,
				"revoker_id": 0,
			}, goqu.C("ends_at").Gt(nowStr)).
			Prepared(true).ToSQL()
		if Sqlx.Get(&draqunCount, draqunSql, draqunArgs...) == nil && draqunCount > 0 {
			return true
		}
	}

	count := uint64(0)
	sql, args, err := Goqu.From(TableBans).
		Select(goqu.COUNT("*")).
		Where(goqu.Ex{
			"group_id":   group.Id,
			"target_id":  userId,
			"revoker_id": 0,
		}, goqu.C("ends_at").Gt(nowStr)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return false
	}

	err = Sqlx.Get(&count, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return count > 0
}

func (group Group) GetActiveBan(userId uint64) *ActionBan {
	if userId == 0 {
		return nil
	}
	nowStr := time.Now().Format(time.RFC3339)
	ban := ActionBan{}
	sql, args, err := Goqu.From(TableBans).
		Where(
			goqu.Ex{
				"group_id":   group.Id,
				"target_id":  userId,
				"revoker_id": 0,
			},
			goqu.C("ends_at").Gt(nowStr),
		).
		Order(goqu.I("acted_at").Desc()).
		Limit(1).
		Prepared(true).ToSQL()
	if err == nil && Sqlx.Get(&ban, sql, args...) == nil {
		ban.SetFullfilled()
		if !ban.IsEmpty() {
			return &ban
		}
	}

	draqunGroup, err := NewGroupFromName("draqun")
	if err == nil && !draqunGroup.IsEmpty() && draqunGroup.Id != group.Id {
		draqunBan := ActionBan{}
		dSql, dArgs, _ := Goqu.From(TableBans).
			Where(
				goqu.Ex{
					"group_id":   draqunGroup.Id,
					"target_id":  userId,
					"revoker_id": 0,
				},
				goqu.C("ends_at").Gt(nowStr),
			).
			Order(goqu.I("acted_at").Desc()).
			Limit(1).
			Prepared(true).ToSQL()
		if Sqlx.Get(&draqunBan, dSql, dArgs...) == nil {
			draqunBan.SetFullfilled()
			if !draqunBan.IsEmpty() {
				return &draqunBan
			}
		}
	}
	return nil
}

func (group Group) ActiveBans() []ActionBan {
	banList := []ActionBan{}
	sql, args, err := Goqu.From(TableBans).
		Where(goqu.Ex{
			"group_id":   group.Id,
			"revoker_id": 0,
		}, goqu.C("ends_at").Gt(time.Now())).
		Order(goqu.I("acted_at").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return banList
	}

	err = Sqlx.Select(&banList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return banList
}

func (group Group) Bans() []ActionBan {
	banList := []ActionBan{}
	sql, args, err := Goqu.From(TableBans).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("acted_at").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return banList
	}

	err = Sqlx.Select(&banList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return banList
}

func (group Group) Kicks() []ActionKick {
	kickList := []ActionKick{}
	sql, args, err := Goqu.From(TableKicks).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("acted_at").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return kickList
	}

	err = Sqlx.Select(&kickList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return kickList
}

func (group Group) Memberships() []ActionMembership {
	membershipList := []ActionMembership{}
	sql, args, err := Goqu.From(TableMemberships).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("acted_at").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return membershipList
	}

	err = Sqlx.Select(&membershipList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return membershipList
}

func (group Group) RoleActions() []ActionRole {
	actionList := []ActionRole{}
	sql, args, err := Goqu.From(TableRoleActions).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("acted_at").Desc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return actionList
	}

	err = Sqlx.Select(&actionList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return actionList
}

func (group Group) MembersCount() uint64 {
	count := uint64(0)
	sql, args, err := Goqu.Select(goqu.COUNT("*")).From(TableMembers).
		Where(goqu.And(
			goqu.C("group_id").Eq(group.Id),
			goqu.C("is_deleted").Eq(false),
		)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return count
	}

	err = Sqlx.Get(&count, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return count
}

func (group Group) UsersPage(page, limit uint) []User {
	userList := []User{}
	from := (page - 1) * limit

	sql, args, err := Goqu.Select(TableUsers+".*").From(TableMembers).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		Where(goqu.Ex{
			TableMembers + ".group_id":   group.Id,
			TableMembers + ".is_deleted": false,
			TableUsers + ".is_deleted":   false,
		}).
		Order(goqu.I(TableMembers + ".user_id").Asc()).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return userList
	}

	err = Sqlx.Select(&userList, sql, args...)
	if err != nil {
		handleErr(err)
	}
	for i := range userList {
		userList[i].SetFullfilled()
	}

	return userList
}

func SearchAllGroups(query string, limit int) ([]Group, error) {
	if limit <= 0 {
		limit = 50
	}
	queryTrim := strings.TrimSpace(query)
	queryLower := "%" + strings.ToLower(queryTrim) + "%"

	exs := []goqu.Expression{
		goqu.I("is_deleted").Eq(false),
	}

	if queryTrim != "" {
		orExprs := []goqu.Expression{
			goqu.L("LOWER(moniker) LIKE ?", queryLower),
			goqu.L("LOWER(name) LIKE ?", queryLower),
			goqu.L("LOWER(description) LIKE ?", queryLower),
		}
		var parsedId uint64
		if _, err := fmt.Sscanf(queryTrim, "%d", &parsedId); err == nil && parsedId > 0 {
			orExprs = append(orExprs, goqu.I("id").Eq(parsedId))
		}
		exs = append(exs, goqu.Or(orExprs...))
	}

	sql, args, err := Goqu.From(TableGroups).Where(exs...).Order(goqu.I("id").Desc()).Limit(uint(limit)).Prepared(true).ToSQL()
	if err != nil {
		return nil, err
	}

	groups := []Group{}
	err = Sqlx.Select(&groups, sql, args...)
	for i := range groups {
		groups[i].SetFullfilled()
	}
	return groups, err
}

func (group Group) SearchUsers(query string, limit int) ([]User, error) {
	if limit <= 0 {
		limit = 10
	}
	userList := []User{}
	queryTrimmed := strings.TrimSpace(query)
	if queryTrimmed == "" {
		return group.UsersPage(1, uint(limit)), nil
	}
	queryLower := "%" + strings.ToLower(queryTrimmed) + "%"

	sql, args, err := Goqu.Select(TableUsers+".*").From(TableMembers).
		Join(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		Where(
			goqu.Ex{
				TableMembers + ".group_id":   group.Id,
				TableMembers + ".is_deleted": false,
				TableUsers + ".is_deleted":   false,
			},
			goqu.Or(
				goqu.L("LOWER("+TableUsers+".moniker) LIKE ?", queryLower),
				goqu.L("LOWER("+TableUsers+".name) LIKE ?", queryLower),
			),
		).
		Limit(uint(limit)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return userList, err
	}

	err = Sqlx.Select(&userList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return userList, err
}

func (group Group) MessageFirst() *Message {
	message := new(Message)

	sql, args, err := Goqu.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Asc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(message)
	if err != nil {
		handleErr(err)
	}

	return message
}

func (group Group) MessageLast() *Message {
	message := new(Message)

	sql, args, err := Goqu.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Desc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(message)
	if err != nil {
		handleErr(err)
		return message
	}

	return message
}

func (group Group) MessagesPage(page, limit uint) []Message {
	messageList := []Message{}
	from := (page - 1) * limit

	subquery := Goqu.From(TableMessages).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("id").Desc()).
		Limit(limit).Offset(from)
	sql, args, err := Goqu.From(subquery).Order(goqu.I("id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return messageList
	}

	err = Sqlx.Select(&messageList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return messageList
}

func (group Group) ActionListPage(page uint, limit uint) []Action {
	actions := []Action{}
	from := (page - 1) * limit
	filter := goqu.Ex{"group_id": group.Id}
	sql, args, err := Goqu.From(TableBans).UnionAll(
		Goqu.From(TableKicks).UnionAll(
			Goqu.From(TableMemberships).Where(filter),
		).Where(filter),
	).Where(filter).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return actions
	}

	err = Sqlx.Select(&actions, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return actions
}

func (group Group) Url(ctx fiber.Ctx) string {
	url, err := ctx.GetRouteURL(
		"page.group", fiber.Map{"group_id": group.Id},
	)
	if err != nil {
		handleErr(err)
		return url
	}

	return url
}

func (group Group) AvatarUrl() string {
	if group.Avatar != "" {
		return string(group.Avatar)
	}
	seed := string(group.Moniker)
	if seed == "" {
		seed = string(group.Name)
	}
	if seed == "" {
		seed = "group"
	}
	return fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=ffffff&color=000000&bold=true", url.QueryEscape(seed))
}

func (group Group) UrlJoin(ctx fiber.Ctx) string {
	url, err := ctx.GetRouteURL(
		"page.group.join", fiber.Map{"group_name": group.Name},
	)
	if err != nil {
		handleErr(err)
		return url
	}

	return url
}
