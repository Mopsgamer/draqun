package connections

import (
	"slices"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v3/log"
)

func New[T comparable]() *Connections[T] {
	return &Connections[T]{
		Users:  &map[uint64][]T{},
		mUsers: &sync.Mutex{},
	}
}

type Connections[T comparable] struct {
	mUsers *sync.Mutex
	// A websocket connection list for each user id.
	Users *map[uint64][]T
}

func (cons Connections[T]) UserConnect(userId uint64, r T) {
	cons.mUsers.Lock()
	(*cons.Users)[userId] = append((*cons.Users)[userId], r)
	cons.mUsers.Unlock()
}

func (cons Connections[T]) UserDisconnect(userId uint64, r T) {
	cons.mUsers.Lock()
	i := slices.Index((*cons.Users)[userId], r)
	(*cons.Users)[userId] = slices.Delete((*cons.Users)[userId], i, i+1)
	cons.mUsers.Unlock()
	log.Info("Closed connection for user " + strconv.FormatUint(userId, 10))
}
