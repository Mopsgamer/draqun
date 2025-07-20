package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

type Group struct {
	Db *goqu.Database
	*model_database.Group
}

func NewGroup(db *goqu.Database) Group {
	return Group{Db: db}
}

func (group *Group) IsEmpty() bool {
	return group.Id != 0 && group.Name != ""
}

// Create new DB record.
func (group *Group) Insert() bool {
	id := Insert(group.Db, TableGroups, group)
	group.Group.Id = *id
	return id != nil
}

// Change the existing DB record.
func (group *Group) Update() bool {
	return Update(group.Db, TableGroups, group, exp.Ex{"id": group.Id})
}

// Delete the existing DB record and memberships.
func (group *Group) Delete() bool {
	deleted := false
	deleted = Delete(group.Db, TableGroups, goqu.Ex{"id": group.Id})
	if !deleted {
		return deleted
	}

	deleted = Delete(group.Db, TableMembers, goqu.Ex{"group_id": group.Id})
	if !deleted {
		return deleted
	}

	// TODO: delete roles(!) and messages(?) when deleting group

	return true
}

// Get the group by her identificator.
func (group *Group) FromId(id uint64) bool {
	group.Group = First[model_database.Group](group.Db, TableGroups, goqu.Ex{"id": id})
	return group.IsEmpty()
}

// Get the group by her group name.
func (group *Group) FromName(name string) bool {
	group.Group = First[model_database.Group](group.Db, TableGroups, goqu.Ex{"name": name})
	return group.IsEmpty()
}

// Get the original group creator.
func (group *Group) Creator() User {
	user := User{Db: group.Db}
	user.FromId(group.CreatorId)
	return user
}

func (group *Group) Users() []User {
	memberList := new([]User)

	err := group.Db.From(TableMembers).Select(TableUsers+".*").Prepared(true).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		ScanStructs(memberList)

	if err == sql.ErrNoRows {
		return *memberList
	}

	if err != nil {
		log.Error(err)
	}

	return *memberList
}

func (group *Group) UsersPage(page, perPage uint) []User {
	memberList := new([]User)
	from := (page - 1) * perPage

	err := group.Db.Select(TableUsers+".*").From(TableMembers).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
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

func (group *Group) MessageFirst() *Message {
	message := new(Message)

	found, err := group.Db.From(TableMessages).Prepared(true).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Asc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (group *Group) MessageLast() *Message {
	message := new(Message)

	found, err := group.Db.From(TableMessages).Prepared(true).Where(goqu.C("group_id").Eq(group.Id)).Order(goqu.C("id").Desc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (group *Group) MessagesPage(page, perPage uint) []Message {
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
