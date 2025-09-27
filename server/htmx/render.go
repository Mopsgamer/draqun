package htmx

import (
	"maps"

	"github.com/gofiber/fiber/v3"
)

func TryRenderPage(ctx fiber.Ctx, name string, bind fiber.Map, layouts ...string) error {
	err := ctx.Render(name, bind, layouts...)
	maps.Copy(bind, fiber.Map{
		"Title":         "500",
		"StatusCode":    fiber.StatusInternalServerError,
		"StatusMessage": fiber.ErrInternalServerError.Message,
		"CenterContent": true,
	})
	if err != nil {
		ctx.Render(
			"partials/x",
			bind,
			"partials/main",
		)
		return err
	}
	return nil
}
