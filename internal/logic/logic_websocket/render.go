package logic_websocket

import (
	"restapp/internal/logic"
	"restapp/websocket"
)

func (ws *LogicWebsocket) SendBytes(message []byte) error {
	return ws.Ctx.WriteMessage(websocket.TextMessage, message)
}

func (ws *LogicWebsocket) SendString(message string) error {
	return ws.SendBytes([]byte(message))
}

func (ws *LogicWebsocket) RenderString(template string, bind any) (string, error) {
	return logic.RenderString(ws.App, template, bind)
}

// Redner template end send websocket message of the result.
func (ws *LogicWebsocket) SendRender(template string, bind any) error {
	str, err := ws.RenderString(template, bind)
	if err != nil {
		return err
	}
	return ws.SendString(str)
}

// Append new flushing data.
func (ws *LogicWebsocket) Push(data string) {
	ws.dataToFlush = ws.dataToFlush + data
}

// Can flush empty string for HTMX requests, it's normal.
func (ws *LogicWebsocket) Flush() error {
	err := ws.SendString(ws.dataToFlush)
	if err != nil {
		return err
	}
	ws.dataToFlush = ""
	return err
}

func wrapSendNotice(ws *LogicWebsocket, message, id string) error {
	return ws.SendString(logic.WrapOob(
		"innerHTML:#"+id,
		&message,
	))
}

func (ws *LogicWebsocket) SendDanger(message, id string) error {
	return wrapSendNotice(ws, message, id)
}

func (ws *LogicWebsocket) SendWarning(message, id string) error {
	return wrapSendNotice(ws, message, id)
}

func (ws *LogicWebsocket) SendSuccess(message, id string) error {
	return wrapSendNotice(ws, message, id)
}
