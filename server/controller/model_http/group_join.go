package model_http

import (
	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/i18n"
)

type GroupJoin struct {
	MemberOfUriGroup
}

func (request *GroupJoin) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "group-join-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	member, user, group := request.Member(ctl)

	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if group == nil {
		return ctl.RenderDanger(i18n.MessageErrGroupNotFound, id)
	}

	if member != nil {
		return ctl.RenderDanger(i18n.MessageErrAlreadyGroupMember, id)
	}

	member = &model_database.Member{
		GroupId:  group.Id,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  false,
		IsBanned: false,
	}

	if !ctl.DB.UserJoinGroup(*member) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	if len(ctl.DB.MemberRoleList(group.Id, user.Id)) < 1 {
		right := model_database.RoleDefault
		rightId := ctl.DB.RoleCreate(right)
		if rightId == nil {
			return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
		}
		right.Id = *rightId

		role := model_database.RoleAssign{
			GroupId: group.Id,
			UserId:  user.Id,
			RightId: right.Id,
		}

		if !ctl.DB.RoleAssign(role) {
			return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
		}
	}

	ctl.HTMXRedirect(controller.PathRedirectGroup(request.GroupId))
	return ctl.RenderSuccess(i18n.MessageSuccJoinedGroup, id)
}
