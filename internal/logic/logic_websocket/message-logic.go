package logic_websocket

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strconv"
)

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
	if ws.DB.GroupMemberById(message.GroupId, message.AuthorId) == nil {
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
