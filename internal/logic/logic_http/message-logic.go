package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic/logic_websocket"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const MessagesPagination uint64 = 5

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
	if r.DB.MemberById(message.GroupId, user.Id) == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	if !model_database.IsValidMessageContent(message.Content) {
		return r.Ctx.SendString(i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString)
	}

	// FIXME: user should have read permissions

	messageId := logic_websocket.MessageSend(*r.DB, *message)
	if messageId == nil {
		return r.Ctx.SendString(i18n.MessageFatalDatabaseQuery)
	}

	message.Id = *messageId
	str, err := r.RenderString("partials/chat-messages", r.MapPage(&fiber.Map{
		"MessageList":        []model_database.Message{*message},
		"MessagesPagination": 2, // disables scroll loading
	}))
	if err != nil {
		return err
	}

	return r.Ctx.SendString(str)
}

func (r LogicHTTP) MessagesPage() error {
	req := new(model_request.MessagesPage)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.Ctx.SendString(i18n.MessageErrInvalidRequest)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.MemberById(*req.GroupId, user.Id) == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	messageList := r.DB.MessageListPage(*req.GroupId, req.Page, MessagesPagination)
	str, err := r.RenderString("partials/chat-messages", fiber.Map{
		"GroupId":            req.GroupId,
		"MessageList":        messageList,
		"MessagesPageNext":   req.Page + 1,
		"MessagesPagination": MessagesPagination,
	})
	if err != nil {
		return err
	}
	return r.Ctx.SendString(str)
}
