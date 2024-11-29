package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic/logic_websocket"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strconv"
)

func (r LogicHTTP) MessageCreate() error {
	req := new(model_request.MessageCreate)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.Ctx.SendString(i18n.MessageErrInvalidRequest)
	}
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.Ctx.SendString(i18n.MessageErrInvalidRequest)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	message := req.Message(user.Id)
	if r.DB.GroupMemberById(message.GroupId, message.AuthorId) == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	if !model_database.IsValidMessageContent(message.Content) {
		return r.Ctx.SendString(i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString)
	}

	// FIXME: user should be a member and have read permissions

	messageId := logic_websocket.MessageSend(*r.DB, *message)
	if messageId == nil {
		return r.Ctx.SendString(i18n.MessageFatalDatabaseQuery)
	}

	return r.Ctx.SendString("")
}
