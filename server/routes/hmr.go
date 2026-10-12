package routes

import (
	"bufio"
	"fmt"
	"sync"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
)

type hmrClient chan string

type hmrHub struct {
	mu        sync.RWMutex
	clients   map[hmrClient]bool
	sessionID string
}

var hub = &hmrHub{
	clients:   make(map[hmrClient]bool),
	sessionID: fmt.Sprintf("%d", time.Now().UnixNano()),
}

func (h *hmrHub) register(c hmrClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
}

func (h *hmrHub) unregister(c hmrClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c)
	}
}

func (h *hmrHub) broadcast(event string, data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, data)
	for client := range h.clients {
		select {
		case client <- msg:
		default:
		}
	}
}

func RouteHMR(app *fiber.App) {
	if environment.BuildEnvironment != environment.BuildModeDevelopment {
		return
	}

	app.Get("/__hmr", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		return c.SendStreamWriter(func(w *bufio.Writer) {
			client := make(hmrClient, 10)
			hub.register(client)
			defer hub.unregister(client)

			initMsg := fmt.Sprintf("event: connected\ndata: {\"sessionId\":\"%s\"}\n\n", hub.sessionID)
			if _, err := w.WriteString(initMsg); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}

			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()

			done := make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						if _, err := w.WriteString(": ping\n\n"); err != nil {
							return
						}
						if err := w.Flush(); err != nil {
							return
						}
					case <-done:
						return
					}
				}
			}()

			for msg := range client {
				if _, err := w.WriteString(msg); err != nil {
					break
				}
				if err := w.Flush(); err != nil {
					break
				}
			}
			close(done)
		})
	})

	app.Post("/__hmr/notify", func(c fiber.Ctx) error {
		eventType := c.Query("type", "reload")
		hub.broadcast(eventType, "{}")
		return c.SendStatus(fiber.StatusOK)
	})
}
