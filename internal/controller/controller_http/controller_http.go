package controller_http

import (
	"restapp/internal/controller/database"

	"github.com/gofiber/fiber/v3"
)

type ControllerHttp struct {
	Ctx fiber.Ctx
	DB  database.Database
}

type Response interface {
	HandleHtmx(ctl ControllerHttp) error
}
