package session

import (
	"slices"
	"sync"
)

type userSessionMap map[uint64][]*ControllerWs

var UserSessionMap = make(userSessionMap)
var sessionMu sync.RWMutex

// Push data for each connection by user id.
func (conns *userSessionMap) Push(data string, pick EventPick) {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	for userId := range *conns {
		for _, ws := range (*conns)[userId] {
			if slices.Contains(ws.Subs, pick) {
				continue
			}
			ws.Push(data)
		}
	}
}

func (conns userSessionMap) Connections(userId uint64) []*ControllerWs {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return conns[userId]
}

func (conns userSessionMap) Connect(userId uint64, ws *ControllerWs) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	conns[userId] = append(conns[userId], ws)
}

func (conns userSessionMap) Close(userId uint64, ws *ControllerWs) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	i := slices.Index(conns[userId], ws)
	if i != -1 {
		conns[userId] = slices.Delete(conns[userId], i, i+1)
	}
}
