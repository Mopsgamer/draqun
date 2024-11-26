package internal

import (
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
)

type Responder struct {
	Ctx fiber.Ctx
	DB  Database
}

type ResponderWS struct {
	Responder
	WS websocket.Conn
}
