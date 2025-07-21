package database

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx/types"
)

type GroupMode string

const (
	GroupModeDm      GroupMode = "dm"
	GroupModePrivate GroupMode = "private"
	GroupModePublic  GroupMode = "public"
)

type Group struct {
	Db *goqu.Database

	Id          uint64        `db:"id"`
	CreatorId   uint64        `db:"creator_id"`
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
	db *goqu.Database,
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
		IsDeleted:   types.BitBool(false),
	}
}

func NewGroupFromId(db *goqu.Database, id uint64) (bool, Group) {
	group := Group{Db: db}
	return group.FromId(id), group
}

func NewGroupFromName(db *goqu.Database, name string) (bool, Group) {
	group := Group{Db: db}
	return group.FromName(name), group
}

func (group Group) IsEmpty() bool {
	return group.Id != 0 && group.Name != ""
}

func (group *Group) Insert() bool {
	id := Insert(group.Db, TableGroups, group)
	group.Id = *id
	return id != nil
}

func (group Group) Update() bool {
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

func (group Group) Everyone() Role {
	role := NewRoleEveryone(group.Db, group.Id)
	role.FromName(role.Name, role.GroupId)
	return role
}

func (group Group) MembersCount() uint64 {
	count := new(uint64)
	ok, err := goqu.Select(goqu.COUNT("*")).From(TableMembers).
		Where(goqu.Ex{TableMembers + ".group_id": group.Id, TableMembers + ".is_deleted": types.BitBool(false)}).
		ScanVal(count)
	if !ok {
		log.Error(err)
	}

	return *count
}

func (group Group) AdminsCount() uint64 {
	count := new(uint64)
	ok, err := goqu.Select(goqu.COUNT("*")).From(TableRoleAssignees).
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(TableRoleAssignees+".role_id"),
			goqu.I(TableRoles+".group_id").Eq(group.Id),
		)).
		Where(goqu.Ex{TableMembers + ".is_deleted": types.BitBool(false)}).
		ScanVal(count)
	if !ok {
		log.Error(err)
	}

	return *count
}

func (group Group) UsersPage(page, perPage uint) []User {
	memberList := new([]User)
	from := (page - 1) * perPage

	err := group.Db.Select(TableUsers+".*").From(TableMembers).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(TableMembers+".user_id"))).
		Where(goqu.Ex{TableMembers + ".group_id": group.Id}).
		Order(goqu.I(TableMembers + ".user_id").Asc()).
		Limit(perPage).Offset(from).
		ScanStructs(memberList)

	if err == sql.ErrNoRows {
		return *memberList
	}

	if err != nil {
		log.Error(err)
	}

	return *memberList
}

func (group Group) MessageFirst() *Message {
	message := new(Message)

	found, err := group.Db.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Asc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (group Group) MessageLast() *Message {
	message := new(Message)

	found, err := group.Db.From(TableMessages).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Desc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (group Group) MessagesPage(page, perPage uint) []Message {
	messageList := new([]Message)
	from := (page - 1) * perPage

	subquery := group.Db.From(TableMessages).
		Where(goqu.Ex{"group_id": group.Id}).
		Order(goqu.I("id").Desc()).
		Limit(perPage).Offset(from)
	err := group.Db.From(subquery.As("subquery")).Order(goqu.I("id").Asc()).
		Executor().ScanStructs(messageList)

	if err == sql.ErrNoRows {
		return *messageList
	}

	if err != nil {
		log.Error(err)
	}

	return *messageList
}
