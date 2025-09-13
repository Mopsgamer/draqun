package htmx

import (
	"github.com/gofiber/fiber/v3"
)

func HandleHTMXError(ctx fiber.Ctx, err error) error {
	level, message := Danger, err.Error()
	if ctx.Get("HX-Error-Wrap") == "false" {
		return ctx.SendString(message)
	}

	if responseErr, ok := err.(Alert); ok {
		level = responseErr.Level()
		message = responseErr.Local()
	}

	bind := fiber.Map{
		"Variant": level.String(),
		"Message": message,
	}

	return ctx.Render("partials/alert", bind)
}
