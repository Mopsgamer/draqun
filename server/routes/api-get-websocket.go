package routes

import (
	"fmt"
	"time"

	"github.com/Mopsgamer/draqun/server/controller_ws"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/websocket"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
)

func RegisterWebsocket(app *fiber.App, db *goqu.Database) {
	logWS := func(start time.Time, err error, ws *controller_ws.ControllerWs) {
		colorErr := fiber.DefaultColors.Green
		if err != nil {
			colorErr = fiber.DefaultColors.Red
		}

		fmt.Printf(
			"%s | %s%3s%s | %13s | %15s | %d | %s%s%s \n",
			time.Now().Format("15:04:05"),
			colorErr,
			"ws",
			fiber.DefaultColors.Reset,
			time.Since(start),
			ws.IP,
			ws.MessageType,
			fiber.DefaultColors.Yellow,
			ws.Message,
			fiber.DefaultColors.Reset,
		)
	}

	// websoket
	app.Get("/groups/:group_id",
		func(ctx fiber.Ctx) error {
			if !websocket.IsWebSocketUpgrade(ctx) {
				return ctx.Next()
			}

			ctxWs := controller_ws.New(ctx)
			user := fiber.Locals[database.User](ctx, perms.LocalAuth)

			return websocket.New(func(conn *websocket.Conn) {
				ctxWs.Conn = conn
				controller_ws.UserSessionMap.Connect(user.Id, ctxWs)
				defer controller_ws.UserSessionMap.Close(user.Id, ctxWs)
				for !ctxWs.Closed {
					messageType, message, err := ctxWs.Conn.ReadMessage()
					if err != nil {
						break
					}

					start := time.Now()
					ctxWs.MessageType = messageType
					ctxWs.Message = message
					err = ctxWs.Flush()

					logWS(start, err, ctxWs)

					if err != nil {
						break
					}
				}
				ctxWs.Closed = true
				ctxWs.Conn.Close()
			})(ctx)
		},
		perms.CheckAuthMember(db, "group_id", func(ctx fiber.Ctx, role database.Role) bool {
			return role.PermMessages.CanReadMessages()
		}),
	)
}
