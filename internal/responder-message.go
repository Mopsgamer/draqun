package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func (r Responder) MessageCreate() error {
	id := "message-send-error"
	req := new(model_request.MessageCreate)
	if err := r.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
	}

	message := req.Message(user.Id)
	if r.DB.GroupMemberById(message.GroupId, message.AuthorId) == nil {
		return r.RenderDanger(MessageErrNotGroupMember, id)
	}

	if !model.IsValidMessageContent(message.Content) {
		return r.RenderWarning(MessageErrMessageContent+" Length: "+strconv.Itoa(len(message.Content))+"/"+strconv.Itoa(model.MessageContentMaxLength), id)
	}

	messageId := r.DB.MessageCreate(*message)
	if messageId == nil {
		return r.RenderDanger(MessageFatalDatabaseQuery, id)
	}

	return r.SendStatus(fiber.StatusOK)
}
