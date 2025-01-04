package controller_http

import (
	"restapp/internal/controller"
	"restapp/internal/controller/controller_ws"
	"restapp/internal/controller/model_database"
	"restapp/internal/controller/model_http"
	"restapp/internal/i18n"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const MessagesPagination uint64 = 5

func (r ControllerHttp) MessageCreate() error {
	id := "error-loading-message-list"
	req := new(model_http.MessageCreate)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderInternalError(id)
	}
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.RenderInternalError(id)
	}

	rights, member, user, group := r.Rights()
	if group == nil {
		return r.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	message := req.Message(user.Id)

	// TODO: ChatWrite should be simple
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

	controller_ws.UserSessionMap.Push(func(userId uint64) bool {
		member := r.DB.MemberById(group.Id, userId)
		if member == nil {
			return false
		}

		if member.IsOwner {
			return true
		}

		rights := r.DB.MemberRights(group.Id, userId)
		return bool(rights.ChatRead)
	}, controller.WrapOob("beforeend:#chat", &str), controller_ws.SubForMessages)

	return r.Ctx.SendString("")
}

func (r ControllerHttp) MessagesPage() error {
	id := "error-loading-message-list"
	req := new(model_http.MessagesPage)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderInternalError(id)
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
