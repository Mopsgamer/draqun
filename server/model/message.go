package model

import (
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Message struct {
	Db *DB `db:"-"`

	Id        uint64    `db:"id"`
	GroupId   uint64    `db:"group_id"`
	AuthorId  uint64    `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func NewMessageFilled(db *DB, groupId, userId uint64, content string) Message {
	return Message{Db: db, GroupId: groupId, AuthorId: userId, Content: strings.TrimSpace(content), CreatedAt: time.Now()}
}

func (message *Message) IsEmpty() bool {
	return message.Id != 0 && message.GroupId != 0 && message.AuthorId != 0
}

func (message *Message) Insert() error {
	return InsertId(message.Db, TableMessages, message, &message.Id)
}

func (message *Message) Update() error {
	return Update(message.Db, TableMessages, message, goqu.Ex{"id": message.Id})
}

func (message *Message) Delete() bool {
	return Delete(message.Db, TableMessages, goqu.Ex{"id": message.Id})
}

func (message *Message) Author() User {
	user := User{Db: message.Db}
	user.FromId(message.AuthorId)
	return user
}
