package logic_websocket

import (
	"bytes"
	"errors"
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

func (r LogicWebsocket) SendContent() error {
	if content := r.GetContent(); content != nil {
		return r.SendString(*content)
	}

	return errors.New("can not send the content. Initialize it: LogicWebsocket")
}

// Create new websocket message as render of a template.
func (r LogicWebsocket) WebsocketRender(template string, bind any) {
	buf, err := logic.RenderBuffer(r.App, template, bind)
	if err != nil {
		log.Error(err)
	}

	r.SendBytesBuffer(buf)
}
