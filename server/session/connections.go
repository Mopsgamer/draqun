package session

import (
	"slices"
	"sync"
	"time"
)

const UserDisconnectGracePeriod = 10 * time.Second

type userSessionMap map[uint64][]*ControllerWs

var UserSessionMap = make(userSessionMap)
var sessionMu sync.RWMutex
var disconnectMu sync.Mutex
var pendingDisconnects = make(map[uint64]*time.Timer)

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

// Push data to all active connections for a specific user id.
func (conns *userSessionMap) PushUser(userId uint64, data string) {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	for _, ws := range (*conns)[userId] {
		ws.Push(data)
	}
}

// Push data to all active connections in a group.
func (conns *userSessionMap) PushGroup(groupId uint64, data string) {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	for userId := range *conns {
		for _, ws := range (*conns)[userId] {
			if ws.GroupId == groupId {
				ws.Push(data)
			}
		}
	}
}

func (conns userSessionMap) Connections(userId uint64) []*ControllerWs {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return conns[userId]
}

// GroupUserIDs returns all unique user IDs that currently have an active connection in the given group.
func (conns *userSessionMap) GroupUserIDs(groupId uint64) []uint64 {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	var userIds []uint64
	for userId := range *conns {
		for _, ws := range (*conns)[userId] {
			if ws.GroupId == groupId {
				userIds = append(userIds, userId)
				break
			}
		}
	}
	return userIds
}

func (conns userSessionMap) IsOnline(userId uint64) bool {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return len(conns[userId]) > 0
}

func (conns userSessionMap) Connect(userId uint64, ws *ControllerWs) {
	disconnectMu.Lock()
	defer disconnectMu.Unlock()
	if timer, ok := pendingDisconnects[userId]; ok {
		timer.Stop()
		delete(pendingDisconnects, userId)
	}

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

func (conns userSessionMap) ScheduleDisconnect(userId uint64, delay time.Duration, onDisconnect func()) {
	disconnectMu.Lock()
	defer disconnectMu.Unlock()

	sessionMu.RLock()
	online := len(conns[userId]) > 0
	sessionMu.RUnlock()
	if online {
		return
	}

	if timer, ok := pendingDisconnects[userId]; ok {
		timer.Stop()
	}
	var timer *time.Timer
	timer = time.AfterFunc(delay, func() {
		disconnectMu.Lock()
		defer disconnectMu.Unlock()

		if pendingDisconnects[userId] != timer {
			return
		}
		delete(pendingDisconnects, userId)

		sessionMu.RLock()
		online := len(conns[userId]) > 0
		sessionMu.RUnlock()
		if !online {
			onDisconnect()
		}
	})
	pendingDisconnects[userId] = timer
}
