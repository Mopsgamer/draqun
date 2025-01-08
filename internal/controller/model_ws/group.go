package model_ws

import (
	"restapp/internal/controller/controller_ws"
	"restapp/internal/i18n"
)

type WebsocketGroup struct {
	Request
	MemberOfUriGroup
}

func (request *WebsocketGroup) HandleHtmx(ctl *controller_ws.ControllerWs) error {
	id := "send-message-error"

	rights, member, group, user := request.Rights(ctl)
	if user == nil {
		return ctl.SendDanger(i18n.MessageErrUserNotFound, id)
	}

	if group == nil {
		return ctl.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	if member == nil {
		return ctl.SendDanger(i18n.MessageErrNotGroupMember, id)
	}

	// TODO: ChatWrite should be simple
	if !member.IsOwner {
		if member.IsBanned {
			return ctl.SendString(i18n.MessageErrNoRights)
		}
		if !rights.ChatRead || !rights.ChatWrite {
			return ctl.SendString(i18n.MessageErrNoRights)
		}
	}

	return ctl.Flush()
}
