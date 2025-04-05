package controller_ws

import (
	"encoding/json"

	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
)

type ControllerWs struct {
	Conn *websocket.Conn
	DB   database.Database
	App  *fiber.App
	IP   string

	User   *model_database.User
	Group  *model_database.Group
	Member *model_database.Member
	Rights *model_database.Role

	MessageType int
	Message     []byte
	dataToFlush string
	Closed      bool
	Subs        []Subscription
}

type Response interface {
	HandleHtmx(ctl *ControllerWs) error
}

func New(ctlHttp controller_http.ControllerHttp) *ControllerWs {
	ws := ControllerWs{
		DB:  ctlHttp.DB,
		App: ctlHttp.Ctx.App(),
		IP:  ctlHttp.Ctx.IP(),

		User:   ctlHttp.User,
		Group:  ctlHttp.Group,
		Member: ctlHttp.Member,
		Rights: ctlHttp.Rights,
	}

	return &ws
}

func (ws *ControllerWs) GetMessageString() string {
	return string(ws.Message)
}

func (ws *ControllerWs) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
