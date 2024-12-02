package logic_websocket

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strconv"
)

func (ws LogicWebsocket) UpdateMessages() error {
	id := "ws-error"
	wsping := new(model_request.WebsocketUpdateChat)
	ws.GetMessageJSON(wsping)

	group := ws.Group()
	if group == nil {
		return ws.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	messageList := ws.DB.MessageListAround(group.Id, wsping.MessageId, 30)
	return ws.SendString(logic.WrapOob("innerHTML:#chat", ws.RenderString("partials/message", messageList)))
}

func (ws LogicWebsocket) UpdateMembers() error {
	id := "ws-error"
	wsping := new(model_request.WebsocketUpdateMembers)
	ws.GetMessageJSON(wsping)

	group := ws.Group()
	if group == nil {
		return ws.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	memberList := ws.DB.MemberListAround(group.Id, wsping.MemberId, 30)
	return ws.SendString(logic.WrapOob("innerHTML:#chat-sidebar", ws.RenderString("partials/chat-messages", memberList)))
}

// Create new chat message, make update events and send websocket message with new chat content. Author is current websocket client.
func (ws LogicWebsocket) MessageCreate() error {
	id := "ws-error"
	req := new(model_request.MessageCreate)
	if err := ws.GetMessageJSON(req); err != nil {
		return ws.SendDanger(i18n.MessageErrInvalidRequest, id)
	}

	user := ws.User()
	if user == nil {
		return ws.SendDanger(i18n.MessageErrUserNotFound, id)
	}

	group := ws.Group()
	if group == nil {
		return ws.SendDanger(i18n.MessageErrGroupNotFound, id)
	}

	message := req.Message(user.Id)
	message.GroupId = group.Id
	if ws.DB.MemberById(message.GroupId, message.AuthorId) == nil {
		return ws.SendDanger(i18n.MessageErrNotGroupMember, id)
	}

	if !model_database.IsValidMessageContent(message.Content) {
		detail := "Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString
		return ws.SendDanger(i18n.MessageErrMessageContent+" "+detail, id)
	}

	// FIXME: user should be a member and have read permissions

	messageId := ws.MessageSend(*message)
	if messageId == nil {
		return ws.SendDanger(i18n.MessageFatalDatabaseQuery, id)
	}
	message.Id = *messageId

	return ws.SendString(logic.WrapOob("beforeend:#chat", ws.RenderString("partials/message", message)))
}
