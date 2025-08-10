package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type User struct {
	Db *DB `db:"-"`

	Id         uint64         `db:"id"`
	Moniker    Moniker        `db:"moniker"` // Nick is customizable name. Can contain emojis and special characters.
	Name       Name           `db:"name"`    // Name is a simple identificator, which can be used to create friend links.
	Email      Email          `db:"email"`
	Phone      Phone          `db:"phone"`
	Password   PasswordHashed `db:"password"`
	Avatar     Avatar         `db:"avatar"`
	CreatedAt  TimePast       `db:"created_at"`
	LastSeenAt TimePast       `db:"last_seen_at"`
	IsDeleted  types.BitBool  `db:"is_deleted"`
}

var _ Model = (*User)(nil)

func NewUser(
	db *DB,
	moniker Moniker,
	name Name,
	email Email,
	phone Phone,
	password PasswordHashed,
	avatar Avatar,
) User {
	return User{
		Db:         db,
		Moniker:    moniker,
		Name:       name,
		Email:      email,
		Phone:      phone,
		Password:   password,
		Avatar:     avatar,
		CreatedAt:  TimePast(time.Now()),
		LastSeenAt: TimePast(time.Now()),
	}
}

func NewUserFromId(db *DB, userId uint64) (User, error) {
	user := User{Db: db}
	err := user.FromId(userId)
	return user, err
}

func NewUserFromEmail(db *DB, email Email) (User, error) {
	user := User{Db: db}
	err := user.FromEmail(email)
	return user, err
}

func NewUserFromName(db *DB, name Name) (User, error) {
	user := User{Db: db}
	err := user.FromName(name)
	return user, err
}

func (user User) IsValid() htmx.Alert {
	if !user.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !user.Name.IsValid() {
		return htmx.AlertFormatName
	}
	if !user.Email.IsValid() {
		return htmx.AlertFormatEmail
	}
	if !user.Phone.IsValid() {
		return htmx.AlertFormatPhone
	}
	if !user.Password.IsValid() {
		return htmx.AlertFormatPassword
	}
	if !user.Avatar.IsValid() {
		return htmx.AlertFormatAvatar
	}
	if !user.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}
	if !user.LastSeenAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}
	return nil
}

func (user User) IsEmpty() bool {
	return user.Id == 0 || user.Name == ""
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

func (user *User) FromEmail(email Email) error {
	return First(user.Db, TableUsers, goqu.Ex{"email": email}, user)
}

func (user *User) FromName(name Name) error {
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
