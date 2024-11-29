package logic_websocket

import (
	"bytes"
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3/log"
)

func (r LogicWebsocket) SendBytes(message []byte) error {
	err := r.Ctx.WriteMessage(websocket.TextMessage, message)
	return err
}

func (r LogicWebsocket) SendString(message string) error {
	return r.SendBytes([]byte(message))
}

func (r LogicWebsocket) SendBytesBuffer(buf bytes.Buffer) error {
	return r.SendString(buf.String())
}

// Redner template end send websocket message of the result.
func (r LogicWebsocket) SendRender(template string, bind any) error {
	buf, err := logic.RenderBuffer(r.App, template, bind)
	if err != nil {
		log.Error(err)
	}

	r.SendBytesBuffer(buf)
	return nil
}
