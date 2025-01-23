package controller_http

import (
	"github.com/Mopsgamer/vibely/internal/controller/database"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"

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
