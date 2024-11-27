package internal

import "restapp/websocket"

type ResponderWebsocket struct {
	Responder
	WS websocket.Conn
}
