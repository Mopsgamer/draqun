package controller_ws

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
	dataToFlush string
	Closed      bool
	Subs        []Subscription
}

func New(ctx fiber.Ctx) *ControllerWs {
	ws := ControllerWs{
		App: ctx.App(),
		IP:  ctx.IP(),
	}

	return &ws
}

func (ws *ControllerWs) GetMessageString() string {
	return string(ws.Message)
}

func (ws *ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
