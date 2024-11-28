package logic_http

import (
	i18n "restapp/internal/i18n"
	"restapp/internal/logic/model"
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

	if !model.IsValidMessageContent(message.Content) {
		return r.Ctx.SendString(i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model.ContentMaxLengthString)
	}

	messageId := r.MessageSend(*message)
	if messageId == nil {
		return r.Ctx.SendString(i18n.MessageFatalDatabaseQuery)
	}

	return r.Ctx.SendString("")
}
