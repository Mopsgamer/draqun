package logic_websocket

import (
	"slices"
	"sync"
)

const (
	SubForMessages string = "messages"
)

var WebsocketConnections = Connections{
	mp:    &map[uint64][]*LogicWebsocket{},
	mutex: &sync.Mutex{},
}

type Connections struct {
	mutex *sync.Mutex
	// A websocket connection list for each user id.
	mp *map[uint64][]*LogicWebsocket
}

// Push data for each connection by user id.
func (conns *Connections) Push(userId uint64, data string, sub string) {
	conns.mutex.Lock()
	for _, ws := range (*conns.mp)[userId] {
		if slices.Contains(ws.Subs, sub) {
			continue
		}
		ws.Push(data)
	}
	conns.mutex.Unlock()
}

func (conns *Connections) Connect(userId uint64, ws *LogicWebsocket) {
	conns.mutex.Lock()
	(*conns.mp)[userId] = append((*conns.mp)[userId], ws)
	conns.mutex.Unlock()
}

func (conns *Connections) Close(userId uint64, ws *LogicWebsocket) {
	conns.mutex.Lock()
	i := slices.Index((*conns.mp)[userId], ws)
	(*conns.mp)[userId] = slices.Delete((*conns.mp)[userId], i, i+1)
	conns.mutex.Unlock()
}
