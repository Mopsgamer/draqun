package controller_ws

import (
	"encoding/json"
	"restapp/internal/controller"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

type ControllerWs struct {
	*controller.Controller
	Ctx *websocket.Conn
	App *fiber.App
	IP  string

	MessageType int
	Message     []byte
	dataToFlush string
	Map         *fiber.Map
	Closed      bool
	Subs        []string
}

func New(ctl *controller.Controller, conn *websocket.Conn, app *fiber.App, ip string, fmap *fiber.Map) ControllerWs {
	ws := ControllerWs{
		Controller: ctl,
		Ctx:        conn,
		Closed:     false,
		App:        app,
		IP:         ip,

		Map: fmap,
	}

	return ws
}

func (ws ControllerWs) GetMessageString() string {
	return string(ws.Message)
}

func (ws ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
