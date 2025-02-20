package model_http

import (
	"strconv"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/i18n"

	"github.com/gofiber/fiber/v3"
)

type MessageCreate struct {
	MemberOfUriGroup
	Content string `form:"content"`
}

func (request *MessageCreate) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "error-loading-message-list"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	rights, member, group, user := request.Rights(ctl)
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

	message := &model_database.Message{
		GroupId:   request.GroupId,
		AuthorId:  user.Id,
		Content:   strings.TrimSpace(request.Content),
		CreatedAt: time.Now(),
	}

	// TODO: ChatWrite should be simple
	if !member.IsOwner {
		if member.IsBanned {
			return ctl.Ctx.SendString(i18n.MessageErrNoRights)
		}
		if !rights.ChatRead || !rights.ChatWrite {
			return ctl.Ctx.SendString(i18n.MessageErrNoRights)
		}
	}

	if !model_database.IsValidMessageContent(message.Content) {
		return ctl.Ctx.SendString(i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString)
	}

	messageId := ctl.DB.MessageCreate(*message)
	if messageId == nil {
		return ctl.Ctx.SendString(i18n.MessageFatalDatabaseQuery)
	}

	message.Id = *messageId
	str, _ := ctl.RenderString("partials/chat-messages", ctl.MapPage(&fiber.Map{
		"MessageList": ctl.DB.CachedMessageList([]model_database.Message{*message}),
	}))

	controller_ws.UserSessionMap.Push(
		func(userId uint64) bool {
			member := ctl.DB.MemberById(group.Id, userId)
			if member == nil {
				return false
			}

			if member.IsOwner {
				return true
			}

			rights := ctl.DB.MemberRights(group.Id, userId)
			return bool(rights.ChatRead)
		},
		controller.WrapOob("beforeend:#chat", &str),
		controller_ws.SubForMessages,
	)

	return ctl.Ctx.SendString("")
}
