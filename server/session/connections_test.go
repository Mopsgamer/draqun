package session

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestUserSessionReconnectCancelsPendingDisconnect(t *testing.T) {
	userId := uint64(time.Now().UnixNano())
	first := &ControllerWs{}
	UserSessionMap.Connect(userId, first)
	UserSessionMap.Close(userId, first)

	disconnected := atomic.Bool{}
	UserSessionMap.ScheduleDisconnect(userId, 50*time.Millisecond, func() {
		disconnected.Store(true)
	})

	reconnected := &ControllerWs{}
	UserSessionMap.Connect(userId, reconnected)
	defer UserSessionMap.Close(userId, reconnected)

	time.Sleep(100 * time.Millisecond)
	if disconnected.Load() {
		t.Fatal("disconnect cleanup ran after the user reconnected")
	}
	if !UserSessionMap.IsOnline(userId) {
		t.Fatal("reconnected user should be online")
	}
}

func TestUserSessionDisconnectRunsAfterGracePeriod(t *testing.T) {
	userId := uint64(time.Now().UnixNano())
	disconnected := make(chan struct{})
	UserSessionMap.ScheduleDisconnect(userId, 10*time.Millisecond, func() {
		close(disconnected)
	})

	select {
	case <-disconnected:
	case <-time.After(time.Second):
		t.Fatal("disconnect cleanup did not run after the grace period")
	}
}
