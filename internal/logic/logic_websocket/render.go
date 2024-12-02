package logic_websocket

import (
	"bytes"
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

func (ws LogicWebsocket) SendBytes(message []byte) error {
	return ws.Ctx.WriteMessage(websocket.TextMessage, message)
}

func (ws LogicWebsocket) SendString(message string) error {
	return ws.SendBytes([]byte(message))
}

func (ws LogicWebsocket) SendBytesBuffer(buf bytes.Buffer) error {
	return ws.SendString(buf.String())
}

func (ws LogicWebsocket) RenderString(template string, bind any) *string {
	return logic.RenderString(ws.App, template, bind)
}

// Redner template end send websocket message of the result.
func (ws LogicWebsocket) SendRender(template string, bind any) error {
	str := ws.RenderString(template, bind)
	if str == nil {
		return nil
	}
	return ws.SendString(*str)
}

func wrapSendNotice(ws LogicWebsocket, template, message, id string) error {
	return ws.SendString(logic.WrapOob(
		"outerHTML:#"+id,
		ws.RenderString(template, fiber.Map{
			"Id":      id,
			"Message": message,
		}),
	))
}

func (ws LogicWebsocket) SendDanger(message, id string) error {
	return wrapSendNotice(ws, "partials/danger", message, id)
}

func (ws LogicWebsocket) SendWarning(message, id string) error {
	return wrapSendNotice(ws, "partials/warning", message, id)
}

func (ws LogicWebsocket) SendSuccess(message, id string) error {
	return wrapSendNotice(ws, "partials/success", message, id)
}
