package logic_websocket

import (
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

func New(appLogic *logic.Logic, conn *websocket.Conn, app *fiber.App, ip string, fmap *fiber.Map) LogicWebsocket {
	ws := LogicWebsocket{
		Logic:  appLogic,
		Ctx:    conn,
		Closed: false,
		App:    app,
		IP:     ip,
		Map:    fmap,
	}

	return ws
}

type LogicWebsocket struct {
	*logic.Logic
	Ctx *websocket.Conn
	App *fiber.App
	IP  string
	// You can change it.
	Closed bool

	Map *fiber.Map
}
