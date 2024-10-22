package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func Index(c fiber.Ctx) error {
	return c.SendFile("./web/templates/index.html")
}

func LoginPage(c fiber.Ctx) error {
	return c.SendFile("./web/templates/login.html")
}

func RegisterPage(c fiber.Ctx) error {
	return c.SendFile("./web/templates/register.html")
}
