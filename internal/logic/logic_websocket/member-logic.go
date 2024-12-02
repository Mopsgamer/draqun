package logic_websocket

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_request"
)

func (ws LogicWebsocket) UpdateMembers() error {
	id := "ws-error"
	wsping := new(model_request.WebsocketUpdateMembers)
	ws.GetMessageJSON(wsping)

	group := ws.Group()
	if group == nil {
		return ws.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	memberList := ws.DB.MemberListAround(group.Id, wsping.MemberId, 30)
	return ws.SendString(logic.WrapOob("innerHTML:#chat-sidebar", ws.RenderString("partials/chat-members", memberList)))
}
