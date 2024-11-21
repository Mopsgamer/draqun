package model_request

import (
	"restapp/internal/model"
	"time"
)

const (
	GroupCreateModePublic  = model.GroupModePublic
	GroupCreateModePrivate = model.GroupModePrivate
)

type GroupCreate struct {
	Name        string `form:"new-group-groupname"`
	Nick        string `form:"new-group-nickname"`
	Password    string `form:"new-group-password"`
	Mode        string `form:"new-group-mode"`
	Description string `form:"new-group-description"`
	Avatar      string `form:"new-group-avatar"`
}

func (g GroupCreate) Group(creatorId uint) *model.Group {
	var password *string = nil
	if g.Password == "" {
		password = &g.Password
	}
	return &model.Group{
		CreatorId: creatorId,
		Nick:      g.Nick,
		Name:      g.Name,
		Mode:      g.Mode,
		Password:  password,
		Avatar:    g.Avatar,
		CreatedAt: time.Now(),
	}
}
