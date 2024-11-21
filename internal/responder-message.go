package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"
	"strconv"
)

func (r Responder) MessageCreate() error {
	id := "message-send-error"
	req := new(model_request.MessageCreate)
	if err := r.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderDanger(MessageErrUserNotFound, id)
	}

	message := req.Message(user.Id)
	if _, err := r.DB.GroupMember(message.GroupId, message.AuthorId); err != nil {
		return r.RenderDanger(MessageErrNotGroupMember, id)
	}

	if !model.IsValidMessageContent(message.Content) {
		return r.RenderWarning(MessageErrMessageContent+" Length: "+strconv.Itoa(len(message.Content))+"/"+strconv.Itoa(model.MessageContentMaxLength), id)
	}

	err = r.DB.MessageCreate(*message)
	if err != nil {
		return r.RenderDanger(MessageFatalCanNotCreateMessage, id)
	}
	return nil
}
