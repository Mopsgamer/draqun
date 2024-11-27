package internal

import (
	"restapp/websocket"
)

type ResponderWebsocket struct {
	Responder
	WS websocket.Conn

	Accept func(r ResponderWebsocket, template string, bind any) (bool, error)
}
