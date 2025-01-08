package controller_http

import (
	"restapp/internal/controller/database"
	"restapp/internal/controller/model_database"

	"github.com/gofiber/fiber/v3"
)

type ControllerHttp struct {
	Ctx fiber.Ctx
	DB  database.Database

	User   *model_database.User
	Group  *model_database.Group
	Member *model_database.Member
	Rights *model_database.Role
}

type Response interface {
	HandleHtmx(ctl ControllerHttp) error
}
