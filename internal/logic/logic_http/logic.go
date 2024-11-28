package logic_http

import (
	"restapp/internal/logic"

	"github.com/gofiber/fiber/v3"
)

type LogicHTTP struct {
	logic.Logic
	Ctx fiber.Ctx
}
