package model_request

import (
	"restapp/internal/logic/model_database"
	"strings"
	"time"
)

type MessageCreate struct {
	*WebsocketRequest
	GroupId uint64 `uri:"group_id"`
	Content string `form:"content" json:"Content"`
}

func (m MessageCreate) Message(authorId uint64) *model_database.Message {
	return &model_database.Message{
		GroupId:   m.GroupId,
		AuthorId:  authorId,
		Content:   strings.TrimSpace(m.Content),
		CreatedAt: time.Now(),
	}
}
