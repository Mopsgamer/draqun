package controller_ws

import (
	"encoding/json"
	"restapp/internal/controller/controller_http"
	"restapp/internal/controller/database"
	"restapp/internal/controller/model_database"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

type ControllerWs struct {
	Conn *websocket.Conn
	DB   database.Database
	App  *fiber.App
	IP   string

	User  *model_database.User
	Group *model_database.Group

	MessageType int
	Message     []byte
	dataToFlush string
	Closed      bool
	Subs        []string
}

type Response interface {
	HandleHtmx(ctl ControllerWs) error
}

func New(ctlHttp controller_http.ControllerHttp) *ControllerWs {
	ws := ControllerWs{
		DB:  ctlHttp.DB,
		App: ctlHttp.Ctx.App(),
		IP:  ctlHttp.Ctx.IP(),
	}

	return &ws
}

func (ws ControllerWs) GetMessageString() string {
	return string(ws.Message)
}

func (ws ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
