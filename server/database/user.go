package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

type User struct {
	Db *goqu.Database
	*model_database.User
}

func NewUser(db *goqu.Database) User {
	return User{Db: db}
}

func (user *User) IsEmpty() bool {
	return user.Id != 0
}

// Create new DB record.
func (user *User) Insert() bool {
	id := Insert(user.Db, TableUsers, user)
	user.User.Id = *id
	return id != nil
}

// Change the existing DB record.
func (user *User) Update() bool {
	return Update(user.Db, TableUsers, user, goqu.Ex(exp.Ex{"id": user.Id}))
}

// Delete the existing DB record.
func (user *User) Delete() bool {
	return Delete(user.Db, TableUsers, goqu.Ex(exp.Ex{"id": user.Id}))
}

// Get the user by his identificator.
func (user *User) FromId(userId uint64) bool {
	user.User = First[model_database.User](user.Db, TableUsers, goqu.Ex{"id": userId})
	return user.IsEmpty()
}

// Get the user by his customizable e-mail.
func (user *User) FromEmail(email string) bool {
	user.User = First[model_database.User](user.Db, TableUsers, goqu.Ex{"email": email})
	return user.IsEmpty()
}

// Get the user by his name customizable identificator.
func (user *User) FromName(name string) bool {
	user.User = First[model_database.User](user.Db, TableUsers, goqu.Ex{"name": name})
	return user.IsEmpty()
}

// Get the list of groups that user is a owner of.
func (user *User) Groups() []Group {
	groupList := new([]Group)

	err := user.Db.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".creator_id": TableMembers + ".user_id"}).
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
func (user *User) Member() []Group {
	groupList := new([]Group)

	err := user.Db.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id}).
		ScanStructs(groupList)

	if err == sql.ErrNoRows {
		return *groupList
	}

	if err != nil {
		log.Error(err)
	}

	return *groupList
}
