package controller_ws

import (
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/Mopsgamer/draqun/websocket"
)

// Append new flushing data.
func (ws *ControllerWs) Push(data string) {
	ws.dataToFlush = []byte(string(ws.dataToFlush) + data)
}

// Can flush empty string for HTMX requests, it's normal.
func (ws *ControllerWs) Flush() error {
	err := ws.Conn.WriteMessage(websocket.TextMessage, ws.dataToFlush)
	if err != nil {
		return err
	}
	ws.dataToFlush = []byte("")
	return nil
}

func wrapSendNotice(ws *ControllerWs, message, id string) error {
	return ws.Conn.WriteMessage(websocket.TextMessage, []byte(render.WrapOob(
		"innerHTML:#"+id,
		&message,
	)))
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
