package model_http

import (
	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
	"github.com/Mopsgamer/vibely/internal/i18n"
)

type GroupChange struct {
	MemberOfUriGroup
	Name        string  `form:"name"`
	Nick        string  `form:"nick"`
	Password    *string `form:"password"`
	Mode        string  `form:"mode"`
	Description string  `form:"description"`
	Avatar      string  `form:"avatar"`
}

func (request *GroupChange) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "group-change-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	rights, member, group, user := request.Rights(ctl)

	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if member == nil {
		return ctl.RenderDanger(i18n.MessageErrNotGroupMember, id)
	}

	if !member.IsOwner {
		if !rights.GroupChange {
			return ctl.RenderDanger(i18n.MessageErrNoRights, id)
		}
	}

	hasChanges := request.Nick != group.Nick ||
		group.Name != request.Name ||
		group.Description != request.Description ||
		group.Mode != request.Mode ||
		group.Password != request.Password

	if !hasChanges {
		return ctl.RenderDanger(i18n.MessageErrUselessChange, id)
	}

	if !model_database.IsValidGroupName(request.Name) {
		return ctl.RenderDanger(i18n.MessageErrGroupName, id)
	}

	if !model_database.IsValidGroupNick(request.Nick) {
		return ctl.RenderDanger(i18n.MessageErrGroupNick, id)
	}

	if !model_database.IsValidGroupPassword(request.Password) {
		return ctl.RenderDanger(i18n.MessageErrGroupPassword, id)
	}

	if !model_database.IsValidGroupDescription(request.Description) {
		return ctl.RenderDanger(i18n.MessageErrGroupDescription, id)
	}

	if !model_database.IsValidGroupMode(request.Mode) {
		return ctl.RenderDanger(i18n.MessageErrGroupMode+" Got: '"+request.Mode+"'.", id)
	}

	group.Nick = request.Nick
	group.Name = request.Name
	group.Description = request.Description
	group.Mode = request.Mode
	group.Password = request.Password
	if !ctl.DB.GroupUpdate(*group) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return nil
}
