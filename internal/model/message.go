package model

import (
	"time"
)

const MessageContentMaxLength int = 8000

type Message struct {
	Id        uint      `db:"id"`
	GroupId   uint      `db:"group_id"`
	AuthorId  uint      `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
