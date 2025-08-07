package session

import (
	"slices"
	"sync"
)

type Subscription string

const (
	SubForMessages Subscription = "messages"
)

var UserSessionMap = userSessionMap{
	mp:    &map[uint64][]*ControllerWs{},
	mutex: &sync.Mutex{},
}

type userSessionMap struct {
	mutex *sync.Mutex
	mp    *map[uint64][]*ControllerWs // A websocket connection list for each user id.
}

// Push data for each connection by user id.
func (conns *userSessionMap) Push(data string, sub Subscription) {
	conns.mutex.Lock()
	for userId := range *conns.mp {
		for _, ws := range (*conns.mp)[userId] {
			if slices.Contains(ws.Subs, sub) {
				continue
			}
			ws.Push(data)
		}
	}
	conns.mutex.Unlock()
}

func (conns *userSessionMap) Connections(userId uint64) []*ControllerWs {
	return (*conns.mp)[userId]
}

func (conns *userSessionMap) Connect(userId uint64, ws *ControllerWs) {
	conns.mutex.Lock()
	(*conns.mp)[userId] = append((*conns.mp)[userId], ws)
	conns.mutex.Unlock()
}

func (conns *userSessionMap) Close(userId uint64, ws *ControllerWs) {
	conns.mutex.Lock()
	i := slices.Index((*conns.mp)[userId], ws)
	(*conns.mp)[userId] = slices.Delete((*conns.mp)[userId], i, i+1)
	conns.mutex.Unlock()
}
