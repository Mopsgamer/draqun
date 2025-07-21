package database

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type Message struct {
	Db        *goqu.Database
	Id        uint64    `db:"id"`
	GroupId   uint64    `db:"group_id"`
	AuthorId  uint64    `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func NewMessage(db *goqu.Database) Message {
	return Message{Db: db}
}

func NewMessageFilled(db *goqu.Database, groupId, userId uint64, content string) Message {
	return Message{Db: db, GroupId: groupId, AuthorId: userId, Content: content, CreatedAt: time.Now()}
}

func (message *Message) IsEmpty() bool {
	return message.Id != 0 && message.GroupId != 0 && message.AuthorId != 0
}

func (message *Message) Insert() bool {
	return Insert(message.Db, TableMessages, message) != nil
}

func (message *Message) Update() bool {
	return Update(message.Db, TableMessages, message, goqu.Ex{"id": message.Id})
}

func (message *Message) Delete() bool {
	return Delete(message.Db, TableMessages, goqu.Ex{"id": message.Id})
}

func (message *Message) Author() User {
	user := NewUser(message.Db)
	user.FromId(message.AuthorId)
	return user
}
