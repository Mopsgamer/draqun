package logic_websocket

import (
	i18n "restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// Create new chat message, make update events and send websocket message with new chat content. Author is current websocket client.
func (ws LogicWebsocket) MessageCreate() error {
	req := new(model_request.MessageCreate)
	if err := ws.Ctx.ReadJSON(req); err != nil {
		bind := logic.MapMerge(ws.Map, &fiber.Map{
			"Id":      "ws-error",
			"Message": i18n.MessageErrInvalidRequest,
		})
		ws.SendRender("partials/chat-group", &bind)
		return nil
	}

	user := ws.User()
	if user == nil {
		return nil
	}

	message := req.Message(user.Id)
	if ws.DB.GroupMemberById(message.GroupId, message.AuthorId) == nil {
		bind := logic.MapMerge(ws.Map, &fiber.Map{
			"Id":      "ws-error",
			"Message": i18n.MessageErrNotGroupMember,
		})
		ws.SendRender("partials/chat-group", &bind)
		return nil
	}

	if !model_database.IsValidMessageContent(message.Content) {
		bind := logic.MapMerge(ws.Map, &fiber.Map{
			"Id":      "ws-error",
			"Message": i18n.MessageErrMessageContent + " Length: " + strconv.Itoa(len(message.Content)) + "/" + model_database.ContentMaxLengthString,
		})
		ws.SendRender("partials/chat-group", &bind)
		return nil
	}

	messageId := ws.MessageSend(*message)
	if messageId == nil {
		bind := logic.MapMerge(ws.Map, &fiber.Map{
			"Id":      "ws-error",
			"Message": i18n.MessageFatalDatabaseQuery,
		})
		ws.SendRender("partials/chat-group", &bind)
		return nil
	}

	ws.SendRender("partials/chat-group", ws.Map)
	return nil
}
