package logic_websocket

import (
	"restapp/internal/connections"
	"restapp/internal/logic"
	"restapp/websocket"
)

var WebsocketConnections = connections.New[*LogicWebsocket]()

type LogicWebsocket struct {
	logic.Logic
	Ctx websocket.Conn

	Accept func(r LogicWebsocket, template string, bind any) (bool, error)
	Closed bool
}

func (r *LogicWebsocket) UserLogout() error {
	r.Closed = true
	return nil
}
