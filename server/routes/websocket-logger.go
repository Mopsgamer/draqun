package routes

import (
	"fmt"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/gofiber/fiber/v3"
)

var prevWsMessage = ""

func logWS(start time.Time, err error, ws *session.ControllerWs) {
	if environment.BuildModeValue == environment.BuildModeProduction {
		return
	}

	colorErr := fiber.DefaultColors.Green
	if err != nil {
		colorErr = fiber.DefaultColors.Red
	}

	if prevWsMessage != string(ws.Message) {
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
	prevWsMessage = string(ws.Message)
}
