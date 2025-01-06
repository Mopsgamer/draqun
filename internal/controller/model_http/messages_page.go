package model_http

import (
	"restapp/internal/controller/controller_http"
	"restapp/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

type MessagesPage struct {
	MemberOfUriGroup
	Page uint64 `uri:"messages_page"`
}

const MessagesPagination uint64 = 5

func (request *MessagesPage) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "error-loading-message-list"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	member, group, user := request.Member(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if group == nil {
		return ctl.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
		return ctl.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	messageList := ctl.DB.MessageListPage(request.GroupId, request.Page, MessagesPagination)
	str, _ := ctl.RenderString("partials/chat-messages", ctl.MapPage(&fiber.Map{
		"GroupId":            request.GroupId,
		"MessageList":        ctl.DB.CachedMessageList(messageList),
		"MessagesPageNext":   request.Page + 1,
		"MessagesPagination": MessagesPagination,
	}))
	return ctl.Ctx.SendString(str)
}
