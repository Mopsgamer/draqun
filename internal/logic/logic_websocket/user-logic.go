package logic_websocket

import (
	"restapp/internal/i18n"
	"restapp/internal/logic/model_database"
)

func UserIsOnline(user *model_database.User) bool {
	if user == nil {
		return false
	}

	cons := *WebsocketConnections.mp
	arr, ok := cons[user.Id]
	return ok && len(arr) > 0
}

func (ws *LogicWebsocket) SubscribeGroup() error {
	id := "send-message-error"

	user := ws.User()
	if user == nil {
		return ws.SendDanger(i18n.MessageErrUserNotFound, id)
	}

	group := ws.Group()
	if group == nil {
		return ws.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	if ws.DB.MemberById(group.Id, user.Id) == nil {
		return ws.SendDanger(i18n.MessageErrNotGroupMember, id)
	}

	return ws.Flush()
}
