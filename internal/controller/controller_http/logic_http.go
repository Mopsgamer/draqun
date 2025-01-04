package controller_http

import (
	"restapp/internal/controller"

	"github.com/gofiber/fiber/v3"
)

type ControllerHttp struct {
	*controller.Controller
	Ctx fiber.Ctx
}
