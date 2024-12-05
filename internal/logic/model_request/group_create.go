package model_request

import (
	"restapp/internal/logic/model_database"
	"time"
)

const (
	GroupCreateModePublic  = model_database.GroupModePublic
	GroupCreateModePrivate = model_database.GroupModePrivate
)

type GroupCreate struct {
	Name        string  `form:"name"`
	Nick        string  `form:"nick"`
	Password    *string `form:"password"`
	Mode        string  `form:"mode"`
	Description string  `form:"description"`
	Avatar      string  `form:"avatar"`
}

func (g GroupCreate) Group(creatorId uint64) *model_database.Group {
	return &model_database.Group{
		CreatorId:   creatorId,
		Nick:        g.Nick,
		Name:        g.Name,
		Mode:        g.Mode,
		Description: g.Description,
		Password:    g.Password,
		Avatar:      g.Avatar,
		CreatedAt:   time.Now(),
	}
}
