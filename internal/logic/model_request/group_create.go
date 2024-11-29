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
	Name        string `form:"groupname"`
	Nick        string `form:"groupnick"`
	Password    string `form:"password"`
	Mode        string `form:"mode"`
	Description string `form:"description"`
	Avatar      string `form:"avatar"`
}

func (g GroupCreate) Group(creatorId uint64) *model_database.Group {
	var password *string = nil
	if g.Password == "" {
		password = &g.Password
	}
	return &model_database.Group{
		CreatorId:   creatorId,
		Nick:        g.Nick,
		Name:        g.Name,
		Mode:        g.Mode,
		Description: g.Description,
		Password:    password,
		Avatar:      g.Avatar,
		CreatedAt:   time.Now(),
	}
}
