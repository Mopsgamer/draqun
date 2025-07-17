package controller

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/gofiber/fiber/v3"
)

func HandleHTMXError(ctx fiber.Ctx, err error) error {
	level, message := htmx.Danger, err.Error()
	if responseErr, ok := err.(htmx.HTMXAlert); ok {
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
	buf, _ := RenderBuffer(ctx.App(), "partials/alert", bind)

	return ctx.Send(buf.Bytes())
}
