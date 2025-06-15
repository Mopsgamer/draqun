package controller_ws

import (
	"encoding/json"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
)

type ControllerWs struct {
	Conn *websocket.Conn
	DB   database.Database
	App  *fiber.App
	IP   string

	MessageType int
	Message     []byte
	dataToFlush string
	Closed      bool
	Subs        []Subscription
}

type Handler func(ctl ControllerWs) error

type Response interface {
	HandleHtmx(ctl *ControllerWs) error
}

func New(ctlHttp controller.Controller) *ControllerWs {
	ws := ControllerWs{
		DB:  ctlHttp.DB,
		App: ctlHttp.Ctx.App(),
		IP:  ctlHttp.Ctx.IP(),
	}

	return &ws
}

func (ws *ControllerWs) GetMessageString() string {
	return string(ws.Message)
}

func (ws *ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
