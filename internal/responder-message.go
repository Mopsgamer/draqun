package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"
	"strconv"
)

func (r Responder) MessageCreate() error {
	req := new(model_request.MessageCreate)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.Ctx.SendString(MessageErrInvalidRequest)
	}
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.Ctx.SendString(MessageErrInvalidRequest)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	message := req.Message(user.Id)
	if r.DB.GroupMemberById(message.GroupId, message.AuthorId) == nil {
		return r.Ctx.SendString(MessageErrNotGroupMember)
	}

	if !model.IsValidMessageContent(message.Content) {
		return r.Ctx.SendString(MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model.ContentMaxLengthString)
	}

	messageId := r.MessageSend(*message)
	if messageId == nil {
		return r.Ctx.SendString(MessageFatalDatabaseQuery)
	}

	return r.Ctx.SendString("")
}
