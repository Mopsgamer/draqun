package controller_ws

import (
	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/websocket"
)

func (ws *ControllerWs) SendBytes(message []byte) error {
	return ws.Conn.WriteMessage(websocket.TextMessage, message)
}

func (ws *ControllerWs) SendString(message string) error {
	return ws.SendBytes([]byte(message))
}

func (ws *ControllerWs) RenderString(template string, bind any) (string, error) {
	return controller.RenderString(ws.App, template, bind)
}

// Redner template end send websocket message of the result.
func (ws *ControllerWs) SendRender(template string, bind any) error {
	str, err := ws.RenderString(template, bind)
	if err != nil {
		return err
	}
	return ws.SendString(str)
}

// Append new flushing data.
func (ws *ControllerWs) Push(data string) {
	ws.dataToFlush = ws.dataToFlush + data
}

// Can flush empty string for HTMX requests, it's normal.
func (ws *ControllerWs) Flush() error {
	err := ws.SendString(ws.dataToFlush)
	if err != nil {
		return err
	}
	ws.dataToFlush = ""
	return nil
}

func wrapSendNotice(ws *ControllerWs, message, id string) error {
	return ws.SendString(controller.WrapOob(
		"innerHTML:#"+id,
		&message,
	))
}

func (ws *ControllerWs) SendDanger(message, id string) error {
	return wrapSendNotice(ws, message, id)
}

func (ws *ControllerWs) SendWarning(message, id string) error {
	return wrapSendNotice(ws, message, id)
}

func (ws *ControllerWs) SendSuccess(message, id string) error {
	return wrapSendNotice(ws, message, id)
}
