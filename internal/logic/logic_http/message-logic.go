package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
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

	rights, member, user, group := r.Rights()
	if group == nil {
		return r.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	message := req.Message(user.Id)
	if !member.IsOwner {
		if member.IsBanned {
			return r.Ctx.SendString(i18n.MessageErrNoRights)
		}
		if !rights.ChatRead || !rights.ChatWrite {
			return r.Ctx.SendString(i18n.MessageErrNoRights)
		}
	}

	if !model_database.IsValidMessageContent(message.Content) {
		return r.Ctx.SendString(i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString)
	}

	messageId := r.DB.MessageCreate(*message)
	if messageId == nil {
		return r.Ctx.SendString(i18n.MessageFatalDatabaseQuery)
	}

	message.Id = *messageId
	str, _ := r.RenderString("partials/chat-messages", r.MapPage(&fiber.Map{
		"MessageList": r.CachedMessageList([]model_database.Message{*message}),
	}))

	logic_websocket.WebsocketConnections.Push(func(userId uint64) bool {
		member := r.DB.MemberById(group.Id, userId)
		if member == nil {
			return false
		}

		if member.IsOwner {
			return true
		}

		rights := r.DB.MemberRights(group.Id, userId)
		return bool(rights.ChatRead)
	}, logic.WrapOob("beforeend:#chat", &str), logic_websocket.SubForMessages)

	return r.Ctx.SendString("")
}

func (r LogicHTTP) MessagesPage() error {
	req := new(model_request.MessagesPage)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.Ctx.SendString(i18n.MessageErrInvalidRequest)
	}

	member, _, group := r.Member()
	if group == nil {
		return r.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	messageList := r.DB.MessageListPage(req.GroupId, req.Page, MessagesPagination)
	str, _ := r.RenderString("partials/chat-messages", r.MapPage(&fiber.Map{
		"GroupId":            req.GroupId,
		"MessageList":        r.CachedMessageList(messageList),
		"MessagesPageNext":   req.Page + 1,
		"MessagesPagination": MessagesPagination,
	}))
	return r.Ctx.SendString(str)
}
