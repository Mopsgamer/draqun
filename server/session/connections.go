package session

import (
	"slices"
)

type userSessionMap map[uint64][]*ControllerWs

var UserSessionMap = make(userSessionMap)

// Push data for each connection by user id.
func (conns *userSessionMap) Push(data string, pick EventPick) {
	for userId := range *conns {
		for _, ws := range (*conns)[userId] {
			if slices.Contains(ws.Subs, pick) {
				continue
			}
			go ws.Push(data)
		}
	}
}

func (conns userSessionMap) Connections(userId uint64) []*ControllerWs {
	return conns[userId]
}

func (conns userSessionMap) Connect(userId uint64, ws *ControllerWs) {
	conns[userId] = append(conns[userId], ws)
}

func (conns userSessionMap) Close(userId uint64, ws *ControllerWs) {
	i := slices.Index(conns[userId], ws)
	conns[userId] = slices.Delete(conns[userId], i, i+1)
}
