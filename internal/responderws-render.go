package internal

import (
	"bytes"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func (r ResponderWebsocket) MapWebsocket(bind *fiber.Map) fiber.Map {
	bindx := r.MapPage(bind)
	if user, tokenErr := r.User(); user != nil {
		bindx["User"] = user
	} else if tokenErr != nil {
		bindx["TokenError"] = true
		bindx["Message"] = "Authorization error"
		bindx["Id"] = "local-token-error"
	}
	return bindx
}

// Create new websocket
func (r ResponderWebsocket) WebsocketRender(template string, bind *fiber.Map) {
	buf := bytes.NewBuffer([]byte{})
	r.Ctx.App().Config().Views.Render(buf, template, r.MapWebsocket(bind))
	log.Info(buf.String())
	r.WS.WriteMessage(websocket.TextMessage, buf.Bytes())
}
