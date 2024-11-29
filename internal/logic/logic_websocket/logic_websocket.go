package logic_websocket

import (
	"encoding/json"
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

type LogicWebsocket struct {
	*logic.Logic
	Ctx         *websocket.Conn
	App         *fiber.App
	IP          string
	MessageType int
	Message     []byte
	// You can change it.
	Closed bool

	Map *fiber.Map
}

func New(appLogic *logic.Logic, conn *websocket.Conn, app *fiber.App, ip string, fmap *fiber.Map) LogicWebsocket {
	ws := LogicWebsocket{
		Logic:  appLogic,
		Ctx:    conn,
		Closed: false,
		App:    app,
		IP:     ip,

		Map: fmap,
	}

	return ws
}

func (ws LogicWebsocket) GetMessageString() string {
	return string(ws.Message)
}

func (ws LogicWebsocket) GetMessageJSON(out any) error {
	return json.Unmarshal(ws.Message, out)
}
