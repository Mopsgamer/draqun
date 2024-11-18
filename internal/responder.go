package internal

import "github.com/gofiber/fiber/v3"

type Responder struct {
	fiber.Ctx
	DB Database
}
