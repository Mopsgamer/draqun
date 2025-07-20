package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) UserCreate(user model_database.User) *uint64 {
	return Insert(db, TableUsers, user)
}

// Change the existing DB record.
func (db Database) UserUpdate(user model_database.User) bool {
	return Update(db, TableUsers, user, goqu.Ex(exp.Ex{"id": user.Id}))
}

// Delete the existing DB record.
func (db Database) UserDelete(userId uint64) bool {
	return Delete(db, TableUsers, goqu.Ex(exp.Ex{"id": userId}))
}

// Get the user by his customizable e-mail.
func (db Database) UserByEmail(email string) *model_database.User {
	return First[model_database.User](db, TableUsers, goqu.Ex{"email": email})
}

// Get the user by his identificator.
func (db Database) UserById(userId uint64) *model_database.User {
	return First[model_database.User](db, TableUsers, goqu.Ex{"id": userId})
}

// Get the user by his name customizable identificator.
func (db Database) UserByName(name string) *model_database.User {
	return First[model_database.User](db, TableUsers, goqu.Ex{"name": name})
}

// Get the list of groups that user is a owner of.
func (db Database) UserOwnGroupList(userId uint64) []model_database.Group {
	groupList := new([]model_database.Group)

	err := db.Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": userId, TableGroups + ".creator_id": TableMembers + ".user_id"}).
		ScanStructs(groupList)

	if err == sql.ErrNoRows {
		return *groupList
	}

	if err != nil {
		log.Error(err)
	}

	return *groupList
}

// Get the list of groups user is a member of.
func (db Database) UserGroupList(userId uint64) []model_database.Group {
	groupList := new([]model_database.Group)

	err := db.Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": userId}).
		ScanStructs(groupList)

	if err == sql.ErrNoRows {
		return *groupList
	}

	if err != nil {
		log.Error(err)
	}

	return *groupList
}
