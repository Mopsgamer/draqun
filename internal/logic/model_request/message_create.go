package model_request

import (
	"restapp/internal/logic/model"
	"strings"
	"time"
)

type MessageCreate struct {
	GroupId uint64 `uri:"group_id"`
	Content string `form:"content"`
}

func (m MessageCreate) Message(authorId uint64) *model.Message {
	return &model.Message{
		GroupId:   m.GroupId,
		AuthorId:  authorId,
		Content:   strings.TrimSpace(m.Content),
		CreatedAt: time.Now(),
	}
}
