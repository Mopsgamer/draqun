package database

import (
	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
)

type Message struct {
	Db *goqu.Database
	*model_database.Message
}

func NewMessage(db *goqu.Database) Message {
	return Message{Db: db}
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
