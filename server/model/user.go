package model

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type User struct {
	Db *DB `db:"-"`

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
	db *DB,
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

func NewUserFromId(db *DB, userId uint64) (User, error) {
	user := User{Db: db}
	return user, user.FromId(userId)
}

func NewUserFromEmail(db *DB, email string) (User, error) {
	user := User{Db: db}
	return user, user.FromEmail(email)
}

func NewUserFromName(db *DB, name string) (User, error) {
	user := User{Db: db}
	return user, user.FromName(name)
}

func (user User) IsEmpty() bool {
	return user.Id == 0
}

func (user *User) Insert() error {
	return InsertId(user.Db, TableUsers, user, &user.Id)
}

func (user User) Update() error {
	return Update(user.Db, TableUsers, user, goqu.Ex{"id": user.Id})
}

func (user *User) FromId(userId uint64) error {
	return First(user.Db, TableUsers, goqu.Ex{"id": userId}, user)
}

func (user *User) FromEmail(email string) error {
	return First(user.Db, TableUsers, goqu.Ex{"email": email}, user)
}

func (user *User) FromName(name string) error {
	return First(user.Db, TableUsers, goqu.Ex{"name": name}, user)
}

func (user *User) GroupListCreator() []Group {
	groupList := []Group{}

	sql, args, err := user.Db.Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".creator_id": TableMembers + ".user_id"}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList
	}

	err = user.Db.Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user *User) GroupListOwner() []Group {
	groupList := []Group{}

	sql, args, err := user.Db.Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".owner_id": TableMembers + ".user_id"}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList
	}

	err = user.Db.Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user *User) GroupList() []Group {
	groupList := []Group{}

	sql, args, err := user.Db.Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList

	}

	err = user.Db.Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user *User) MemberList() []Member {
	memberList := []Member{}

	sql, args, err := user.Db.Goqu.Select(TableMembers+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(TableMembers+".group_id"))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return memberList

	}

	err = user.Db.Sqlx.Select(&memberList, sql, args...)
	if err != nil {
		handleErr(err)
		return memberList
	}

	return memberList
}
