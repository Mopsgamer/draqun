package model

import (
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type Message struct {
	Db *DB `db:"-"`

	Id        uint64         `db:"id"`
	GroupId   uint64         `db:"group_id"`
	AuthorId  uint64         `db:"author_id"`
	Content   MessageContent `db:"content"`
	CreatedAt TimePast       `db:"created_at"`
}

var _ Model = (*Message)(nil)

func NewMessageFilled(db *DB, groupId, userId uint64, content string) Message {
	return Message{
		Db:        db,
		GroupId:   groupId,
		AuthorId:  userId,
		Content:   MessageContent(strings.TrimSpace(content)),
		CreatedAt: TimePast(time.Now()),
	}
}

func (message Message) IsValid() htmx.Alert {
	if !message.Content.IsValid() {
		return htmx.AlertFormatMessageContent
	}
	if !message.CreatedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (message Message) IsEmpty() bool {
	return message.GroupId == 0 || message.AuthorId == 0 || message.Content == ""
}

func (message *Message) Insert() error {
	return InsertId(message.Db, TableMessages, message, &message.Id)
}

func (message Message) Update() error {
	return Update(message.Db, TableMessages, message, goqu.Ex{"id": message.Id})
}

func (message Message) Delete() error {
	return Delete(message.Db, TableMessages, goqu.Ex{"id": message.Id})
}

func (message *Message) Author() User {
	user := User{Db: message.Db}
	user.FromId(message.AuthorId)
	return user
}
