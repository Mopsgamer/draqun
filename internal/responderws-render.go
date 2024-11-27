package internal

import (
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
func (r ResponderWebsocket) WebsocketRender(template string, bind any) {
	var bindx any
	if bindMap, ok := bind.(*fiber.Map); ok {
		bindx = r.MapWebsocket(bindMap)
	} else {
		bindx = bind
	}

	accepted, err := r.Accept(r, template, bindx)
	if err != nil {
		log.Error(err)
		buf, err := r.RenderBuffer("partials/danger", fiber.Map{
			"Id":      "ws-err",
			"Message": err.Error(),
		})
		if err != nil {
			log.Error(err)
		}

		r.WS.WriteMessage(websocket.CloseMessage, buf.Bytes())
		// r.WS.Close() // https://github.com/gofiber/contrib/issues/698
		r.Closed = true // workaround
		return
	}
	if !accepted {
		return
	}

	buf, err := r.RenderBuffer(template, bindx)
	if err != nil {
		log.Error(err)
	}
	r.WS.WriteMessage(websocket.TextMessage, buf.Bytes())
}
