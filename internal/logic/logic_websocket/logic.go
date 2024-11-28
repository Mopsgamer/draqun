package logic_websocket

import (
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

func New(appLogic *logic.Logic, conn *websocket.Conn, app *fiber.App, updateContent func(r *LogicWebsocket) *string, bind *fiber.Map) LogicWebsocket {
	r := LogicWebsocket{
		Logic:         appLogic,
		Ctx:           conn,
		Closed:        false,
		App:           app,
		RenderContent: updateContent,
		Bind:          bind,
	}

	r.content = updateContent(&r)
	return r
}

type LogicWebsocket struct {
	*logic.Logic
	Ctx           *websocket.Conn
	App           *fiber.App
	Closed        bool
	RenderContent func(r *LogicWebsocket) *string

	Bind    *fiber.Map
	content *string
}

func (r *LogicWebsocket) UpdateContent() {
	if str := r.RenderContent(r); str != nil {
		r.content = str
		return
	}
}

func (r LogicWebsocket) GetContent() *string {
	return r.content
}
