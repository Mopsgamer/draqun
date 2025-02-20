package model_http

import (
	"github.com/Mopsgamer/draqun/internal/controller/controller_http"
	"github.com/Mopsgamer/draqun/internal/i18n"
)

type GroupLeave struct {
	MemberOfUriGroup
}

func (request *GroupLeave) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "group-leave-error"
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

	// TODO: do not leave if last owner and there are other non-owner members.
	// Ask for assigning new owner before leave.

	// TODO: delete group on leave if no other members.

	ctl.HTMXRedirect("/chat")
	return ctl.RenderSuccess(i18n.MessageSuccLeavedGroup, id)
}
