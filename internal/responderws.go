package internal

import (
	"restapp/internal/connections"
	"restapp/websocket"
)

type ResponderWebsocket struct {
	Responder
	WS websocket.Conn

	Accept func(r ResponderWebsocket, template string, bind any) (bool, error)
	Closed bool
}

var WebsocketConnections = connections.New[*ResponderWebsocket]()
