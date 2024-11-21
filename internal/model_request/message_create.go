package model_request

import (
	"restapp/internal/model"
	"strings"
	"time"
)

type MessageCreate struct {
	GroupId uint   `uri:"group_id"`
	Content string `form:"content"`
}

func (m MessageCreate) Message(authorId uint) *model.Message {
	return &model.Message{
		GroupId:   m.GroupId,
		AuthorId:  authorId,
		Content:   strings.TrimSpace(m.Content),
		CreatedAt: time.Now(),
	}
}
