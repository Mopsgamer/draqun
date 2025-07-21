package database

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

type User struct {
	Db *goqu.Database

	Id         uint64    `db:"id"`
	Moniker    string    `db:"moniker"` // Nick is customizable name. Can contain emojis and special characters.
	Name       string    `db:"name"`    // Name is a simple identificator, which can be used to create friend links.
	Email      string    `db:"email"`
	Phone      string    `db:"phone"`
	Password   string    `db:"password"` // Hashed password string.
	Avatar     string    `db:"avatar"`
	CreatedAt  time.Time `db:"created_at"`
	LastSeenAt time.Time `db:"last_seen_at"`
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
	user.Id = *id
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
	First(user.Db, TableUsers, goqu.Ex{"id": userId}, user)
	return user.IsEmpty()
}

// Get the user by his customizable e-mail.
func (user *User) FromEmail(email string) bool {
	First(user.Db, TableUsers, goqu.Ex{"email": email}, user)
	return user.IsEmpty()
}

// Get the user by his name customizable identificator.
func (user *User) FromName(name string) bool {
	First(user.Db, TableUsers, goqu.Ex{"name": name}, user)
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
