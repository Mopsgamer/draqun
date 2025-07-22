package database

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx/types"
)

type User struct {
	Db *goqu.Database

	Id         uint64        `db:"id"`
	Moniker    string        `db:"moniker"` // Nick is customizable name. Can contain emojis and special characters.
	Name       string        `db:"name"`    // Name is a simple identificator, which can be used to create friend links.
	Email      string        `db:"email"`
	Phone      string        `db:"phone"`
	Password   string        `db:"password"` // Hashed password string.
	Avatar     string        `db:"avatar"`
	CreatedAt  time.Time     `db:"created_at"`
	LastSeenAt time.Time     `db:"last_seen_at"`
	IsDeleted  types.BitBool `db:"is_deleted"`
}

func NewUser(
	db *goqu.Database,
	moniker, name, email, phone, password, avatar string,
) User {
	return User{
		Db:         db,
		Moniker:    moniker,
		Name:       name,
		Email:      email,
		Phone:      phone,
		Password:   password,
		Avatar:     avatar,
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}
}

func NewUserFromId(db *goqu.Database, userId uint64) (bool, User) {
	user := User{Db: db}
	return user.FromId(userId), user
}

func NewUserFromEmail(db *goqu.Database, email string) (bool, User) {
	user := User{Db: db}
	return user.FromEmail(email), user
}

func NewUserFromName(db *goqu.Database, name string) (bool, User) {
	user := User{Db: db}
	return user.FromName(name), user
}

func (user *User) IsEmpty() bool {
	return user.Id != 0
}

func (user *User) Insert() bool {
	id := Insert(user.Db, TableUsers, user)
	user.Id = *id
	return id != nil
}

func (user *User) Update() bool {
	return Update(user.Db, TableUsers, user, goqu.Ex{"id": user.Id})
}

func (user *User) FromId(userId uint64) bool {
	First(user.Db, TableUsers, goqu.Ex{"id": userId}, user)
	return user.IsEmpty()
}

func (user *User) FromEmail(email string) bool {
	First(user.Db, TableUsers, goqu.Ex{"email": email}, user)
	return user.IsEmpty()
}

func (user *User) FromName(name string) bool {
	First(user.Db, TableUsers, goqu.Ex{"name": name}, user)
	return user.IsEmpty()
}

func (user *User) GroupListCreator() []Group {
	groupList := new([]Group)

	err := user.Db.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
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

func (user *User) GroupListOwner() []Group {
	groupList := new([]Group)

	err := user.Db.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".owner_id": TableMembers + ".user_id"}).
		ScanStructs(groupList)

	if err == sql.ErrNoRows {
		return *groupList
	}

	if err != nil {
		log.Error(err)
	}

	return *groupList
}

func (user *User) GroupList() []Group {
	groupList := new([]Group)

	err := user.Db.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
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

func (user *User) MemberList() []Member {
	groupList := new([]Member)

	err := user.Db.Select(TableMembers+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
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
