package model_database

import (
	"time"
)

type Message struct {
	Id        uint64    `db:"id"`
	GroupId   uint64    `db:"group_id"`
	AuthorId  uint64    `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
