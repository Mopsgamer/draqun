package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type User struct {
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
	moniker Moniker,
	name Name,
	email Email,
	phone Phone,
	password PasswordHashed,
	avatar Avatar,
) User {
	return User{
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

func NewUserFromId(userId uint64) (User, error) {
	user := User{}
	return user, First(TableUsers, goqu.Ex{"id": userId}, &user)
}

func NewUserFromEmail(email Email) (User, error) {
	user := User{}
	return user, First(TableUsers, goqu.Ex{"email": email}, &user)
}

func NewUserFromName(name Name) (User, error) {
	user := User{}
	return user, First(TableUsers, goqu.Ex{"name": name}, &user)
}

func (user User) Validate() htmx.Alert {
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
	return InsertId(TableUsers, user, &user.Id)
}

func (user User) Update() error {
	return Update(TableUsers, user, goqu.Ex{"id": user.Id})
}

func (user User) GroupListCreator() []Group {
	groupList := []Group{}

	sql, args, err := Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".creator_id": TableMembers + ".user_id"}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList
	}

	err = Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user User) GroupListOwner() []Group {
	groupList := []Group{}

	sql, args, err := Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id, TableGroups + ".owner_id": TableMembers + ".user_id"}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList
	}

	err = Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user User) GroupList() []Group {
	groupList := []Group{}

	sql, args, err := Goqu.Select(TableGroups+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return groupList
	}

	err = Sqlx.Select(&groupList, sql, args...)
	if err != nil {
		handleErr(err)
		return groupList
	}

	return groupList
}

func (user User) MemberList() []Member {
	memberList := []Member{}

	sql, args, err := Goqu.Select(TableMembers+".*").From(TableGroups).
		LeftJoin(goqu.I(TableMembers), goqu.On(goqu.I(TableGroups+".id").Eq(goqu.I(TableMembers+".group_id")))).
		Where(goqu.Ex{TableMembers + ".user_id": user.Id}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return memberList

	}

	err = Sqlx.Select(&memberList, sql, args...)
	if err != nil {
		handleErr(err)
		return memberList
	}

	return memberList
}
