package logic_websocket

import (
	"slices"
	"sync"
)

var WebsocketConnections = Connections{
	Users:  &map[uint64][]*LogicWebsocket{},
	mUsers: &sync.Mutex{},
}

type Connections struct {
	mUsers *sync.Mutex
	// A websocket connection list for each user id.
	Users *map[uint64][]*LogicWebsocket
}

func (cons Connections) UserUpdateContent(userId uint64) {
	for _, ws := range (*cons.Users)[userId] {
		ws.UpdateContent()
	}
}

func (cons Connections) UserConnect(userId uint64, ws *LogicWebsocket) {
	cons.mUsers.Lock()
	(*cons.Users)[userId] = append((*cons.Users)[userId], ws)
	cons.mUsers.Unlock()
}

func (cons Connections) UserDisconnect(userId uint64, ws *LogicWebsocket) {
	cons.mUsers.Lock()
	i := slices.Index((*cons.Users)[userId], ws)
	(*cons.Users)[userId] = slices.Delete((*cons.Users)[userId], i, i+1)
	cons.mUsers.Unlock()
}
