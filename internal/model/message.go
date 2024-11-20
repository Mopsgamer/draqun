package model

import "time"

type Message struct {
	Id        string    `db:"id"`
	GroupId   string    `db:"group_id"`
	AuthorId  string    `db:"author_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
