package model_request

import (
	"restapp/internal/logic/model_database"
	"strings"
	"time"
)

type MessageCreate struct {
	*GroupUri
	Content string `form:"content"`
}

func (m MessageCreate) Message(authorId uint64) *model_database.Message {
	return &model_database.Message{
		GroupId:   m.GroupId,
		AuthorId:  authorId,
		Content:   strings.TrimSpace(m.Content),
		CreatedAt: time.Now(),
	}
}
