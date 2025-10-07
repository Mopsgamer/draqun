package htmx

import (
	"github.com/gofiber/fiber/v3"
)

func HandleHTMXError(ctx fiber.Ctx, err error) error {
	level, message := Danger, err.Error()

	if responseErr, ok := err.(Alert); ok {
		level = responseErr.Level()
		message = responseErr.Local()
	}

	if ctx.Get("HX-Error-Wrap") == "false" {
		return ctx.SendString(message)
	}

	bind := fiber.Map{
		"Variant": level.String(),
		"Message": message,
	}

	return ctx.Render("partials/alert", bind)
}
