package model

import (
	"time"

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

type Group struct {
	Db *DB `db:"-"`

	Id          uint64        `db:"id"`
	CreatorId   uint64        `db:"creator_id"`
	OwnerId     uint64        `db:"owner_id"`
	Moniker     string        `db:"moniker"` // Nick is a customizable name.
	Name        string        `db:"name"`    // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode     `db:"mode"`
	Password    string        `db:"password"` // Optional hashed password string.
	Description string        `db:"description"`
	Avatar      string        `db:"avatar"`
	CreatedAt   time.Time     `db:"created_at"`
	IsDeleted   types.BitBool `db:"is_deleted"`
}

func NewGroup(
	db *DB,
	creatorId uint64,
	moniker, name string,
	mode GroupMode,
	password, description, avatar string,
) Group {
	return Group{
		Db:          db,
		CreatorId:   creatorId,
		Moniker:     moniker,
		Name:        name,
		Mode:        mode,
		Password:    password,
		Description: description,
		Avatar:      avatar,
		CreatedAt:   time.Now(),
		IsDeleted:   false,
	}
}

func NewGroupFromId(db *DB, id uint64) (bool, Group) {
	group := Group{Db: db}
	return group.FromId(id), group
}

func NewGroupFromName(db *DB, name string) (bool, Group) {
	group := Group{Db: db}
	return group.FromName(name), group
}

func (group Group) IsEmpty() bool {
	return group.Id != 0 && group.Name != ""
}

func (group *Group) Insert() error {
	return InsertId(group.Db, TableGroups, group, &group.Id)
}

func (group Group) Update() error {
	return Update(group.Db, TableGroups, group, goqu.Ex{"id": group.Id})
}

func (group *Group) FromId(id uint64) bool {
	First(group.Db, TableGroups, goqu.Ex{"id": id}, group)
	return group.IsEmpty()
}

func (group *Group) FromName(name string) bool {
	First(group.Db, TableGroups, goqu.Ex{"name": name}, group)
	return group.IsEmpty()
}

func (group Group) Creator() User {
	user := User{Db: group.Db}
	user.FromId(group.CreatorId)
	return user
}

func (group Group) Owner() User {
	user := User{Db: group.Db}
	user.FromId(group.OwnerId)
	return user
}

func (group Group) Everyone() Role {
	role := NewRoleEveryone(group.Db, group.Id)
	role.FromName(role.Name, role.GroupId)
	return role
}

func (group Group) MembersCount() uint64 {
	count := uint64(0)
	sql, args, err := goqu.Select(goqu.COUNT("*")).From(TableMembers).
		Where(goqu.Ex{TableMembers + ".group_id": group.Id, TableMembers + ".is_deleted": types.BitBool(false)}).
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
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(TableMembers+".user_id"))).
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
