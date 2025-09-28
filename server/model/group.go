package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx/types"
)

type GroupMode string

const (
	GroupModeDm      GroupMode = "dm"
	GroupModePrivate GroupMode = "private"
	GroupModePublic  GroupMode = "public"
)

func (gm GroupMode) IsValid() bool {
	return gm == GroupModeDm || gm == GroupModePrivate || gm == GroupModePublic
}

type Group struct {
	Db *DB `db:"-"`

	Id          uint64                 `db:"id"`
	CreatorId   uint64                 `db:"creator_id"`
	OwnerId     uint64                 `db:"owner_id"`
	Moniker     Moniker                `db:"moniker"` // Nick is a customizable name.
	Name        Name                   `db:"name"`    // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode              `db:"mode"`
	Password    OptionalPasswordHashed `db:"password"`
	Description Description            `db:"description"`
	Avatar      Avatar                 `db:"avatar"`
	CreatedAt   TimePast               `db:"created_at"`
	IsDeleted   types.BitBool          `db:"is_deleted"`
}

var _ Model = (*Group)(nil)

func NewGroup(
	db *DB,
	creatorId uint64,
	moniker Moniker,
	name Name,
	mode GroupMode,
	password OptionalPasswordHashed,
	description Description,
	avatar Avatar,
) Group {
	return Group{
		Db:          db,
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

func NewGroupFromId(db *DB, id uint64) (Group, error) {
	group := Group{Db: db}
	err := group.FromId(id)
	return group, err
}

func NewGroupFromName(db *DB, name Name) (Group, error) {
	group := Group{Db: db}
	err := group.FromName(name)
	return group, err
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

func (group Group) IsEmpty() bool {
	return group.Id == 0 || group.Name == ""
}

func (group *Group) Insert() error {
	return InsertId(group.Db, TableGroups, group, &group.Id)
}

func (group Group) Update() error {
	return Update(group.Db, TableGroups, group, goqu.Ex{"id": group.Id})
}

func (group *Group) FromId(id uint64) error {
	return First(group.Db, TableGroups, goqu.Ex{"id": id}, group)
}

func (group *Group) FromName(name Name) error {
	return First(group.Db, TableGroups, goqu.Ex{"name": name}, group)
}

func (group Group) Creator() User {
	user := User{Db: group.Db}
	_ = user.FromId(group.CreatorId)
	return user
}

func (group Group) Owner() User {
	user := User{Db: group.Db}
	_ = user.FromId(group.OwnerId)
	return user
}

func (group Group) Everyone() Role {
	role := NewRoleEveryone(group.Db, group.Id)
	_ = role.FromName(role.Name, role.GroupId)
	return role
}

func (group Group) MembersCount() uint64 {
	count := uint64(0)
	sql, args, err := group.Db.Goqu.Select(goqu.COUNT("*")).From(TableMembers).
		Where(goqu.And(
			goqu.C("group_id").Eq(group.Id),
			goqu.C("is_deleted").Eq([]byte{0}),
		)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return count
	}

	err = group.Db.Sqlx.Get(&count, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return count
}

func (group Group) UsersPage(page, limit uint) []User {
	userList := []User{}
	from := (page - 1) * limit

	sql, args, err := group.Db.Goqu.Select(TableUsers+".*").From(TableMembers).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		Where(goqu.Ex{TableMembers + ".group_id": group.Id}).
		Order(goqu.I(TableMembers + ".user_id").Asc()).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return userList
	}

	err = group.Db.Sqlx.Select(&userList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	for i := range userList {
		userList[i].Db = group.Db
	}

	return userList
}

func (group Group) MessageFirst() *Message {
	message := new(Message)

	sql, args, err := group.Db.Goqu.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Asc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = group.Db.Sqlx.QueryRowx(sql, args...).StructScan(&message)
	if err != nil {
		handleErr(err)
	}

	return message
}

func (group Group) MessageLast() *Message {
	message := new(Message)

	sql, args, err := group.Db.Goqu.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Desc()).Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return message
	}

	err = group.Db.Sqlx.QueryRowx(sql, args...).StructScan(&message)
	if err != nil {
		handleErr(err)
		return message
	}

	return message
}

func (group Group) MessagesPage(page, limit uint) []Message {
	messageList := []Message{}
	from := (page - 1) * limit

	subquery := group.Db.Goqu.From(TableMessages).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("id").Desc()).
		Limit(limit).Offset(from)
	sql, args, err := group.Db.Goqu.From(subquery).Order(goqu.I("id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return messageList
	}

	err = group.Db.Sqlx.Select(&messageList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	for i := range messageList {
		messageList[i].Db = group.Db
	}

	return messageList
}

func (group Group) ActionListPage(page uint, limit uint) []Action {
	actions := []Action{}
	from := (page - 1) * limit
	filter := goqu.Ex{"group_id": group.Id}
	sql, args, err := group.Db.Goqu.From(TableBans).UnionAll(
		group.Db.Goqu.From(TableKicks).UnionAll(
			group.Db.Goqu.From(TableMemberships).Where(filter),
		).Where(filter),
	).Where(filter).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return actions
	}

	err = group.Db.Sqlx.Select(&actions, sql, args...)
	if err != nil {
		handleErr(err)
	}

	for i := range actions {
		actions[i].SetDb(group.Db)
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
