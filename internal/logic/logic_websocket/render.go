package logic_websocket

import (
	"restapp/internal/logic"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func (r LogicWebsocket) MapWebsocket(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}
	if user, tokenErr := r.User(); user != nil {
		bindx["User"] = user
	} else if tokenErr != nil {
		bindx["TokenError"] = true
		bindx["Message"] = "Authorization error"
		bindx["Id"] = "local-token-error"
	}

	bindx = logic.MapMerge(&bindx, bind)
	return bindx
}

// Create new websocket
func (r LogicWebsocket) WebsocketRender(template string, bind any) {
	var bindx any
	if bindMap, ok := bind.(*fiber.Map); ok {
		bindx = r.MapWebsocket(bindMap)
	} else {
		bindx = bind
	}

	accepted, err := r.Accept(r, template, bindx)
	if err != nil {
		log.Error(err)
		message := err.Error()
		buf, renderErr := r.RenderBuffer("partials/danger", fiber.Map{
			"Id":      "ws-err",
			"Message": message,
		})
		if renderErr != nil {
			log.Error(renderErr)
		}

		r.Ctx.WriteMessage(websocket.CloseMessage, buf.Bytes())
		// r.Ctx.Close() // https://github.com/gofiber/contrib/issues/698
		r.Closed = true // workaround
		return
	}
	if !accepted {
		log.Info("not accepted")
		return
	}

	buf, err := r.RenderBuffer(template, bindx)
	if err != nil {
		log.Error(err)
	}
	r.Ctx.WriteMessage(websocket.TextMessage, buf.Bytes())
}
