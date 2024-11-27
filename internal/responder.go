package internal

import (
	"github.com/gofiber/fiber/v3"
)

type Responder struct {
	Ctx fiber.Ctx
	DB  Database
}

func (r Responder) Map(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}
	if bind != nil {
		for k, v := range *bind {
			bindx[k] = v
		}
	}
	return bindx
}
