package model_ws

import (
	"restapp/internal/controller/controller_ws"
	"restapp/internal/controller/model_database"
)

type MemberOfUriGroup struct {
	UriGroupId
	CookieUserToken
}

func (request *MemberOfUriGroup) Member(ctl controller_ws.ControllerWs) (*model_database.Member, *model_database.Group, *model_database.User) {
	user := request.User(ctl)
	group := request.Group(ctl)

	if group != nil && user != nil {
		return ctl.DB.MemberById(group.Id, user.Id), group, user
	}

	return nil, group, user
}

func (request *MemberOfUriGroup) Rights(ctl controller_ws.ControllerWs) (*model_database.Role, *model_database.Member, *model_database.Group, *model_database.User) {
	member, group, user := request.Member(ctl)

	if group != nil && user != nil {
		rights := ctl.DB.MemberRights(request.GroupId, user.Id)
		return &rights, member, group, user
	}

	return nil, member, group, user
}
