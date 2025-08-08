package session

import (
	"encoding/json"

	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
)

type ControllerWs struct {
	Conn *websocket.Conn
	App  *fiber.App
	IP   string

	MessageType int
	Message     []byte
	dataToFlush []byte
	Closed      bool
	Subs        []EventPick
}

func New(ctx fiber.Ctx) *ControllerWs {
	ws := ControllerWs{
		App: ctx.App(),
		IP:  ctx.IP(),
	}

	return &ws
}

func (ws *ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}

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
