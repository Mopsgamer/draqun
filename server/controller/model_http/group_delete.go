package model_http

import (
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/i18n"
)

type GroupDelete struct {
	MemberOfUriGroup
}

func (request *GroupDelete) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "group-delete-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	member, group, user := request.Member(ctl)

	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if group == nil {
		return ctl.RenderDanger(i18n.MessageErrGroupNotFound, id)
	}

	if member == nil {
		return ctl.RenderDanger(i18n.MessageErrNotGroupMember, id)
	}

	if !member.IsOwner {
		return ctl.RenderDanger(i18n.MessageErrNoRights, id)
	}

	if !ctl.DB.GroupDelete(request.GroupId) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return nil
}
