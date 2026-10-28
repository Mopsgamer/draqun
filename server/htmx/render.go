package htmx

import (
	"maps"

	"github.com/gofiber/fiber/v3"
)

var OnRenderError func(name string, err error, path string)

func TryRenderPage(ctx fiber.Ctx, name string, bind fiber.Map, layouts ...string) error {
	err := ctx.Render(name, bind, layouts...)

	if err != nil {
		if OnRenderError != nil {
			OnRenderError(name, err, ctx.Path())
		}

		maps.Copy(bind, fiber.Map{
			"Title":         "500",
			"StatusCode":    fiber.StatusInternalServerError,
			"StatusMessage": fiber.ErrInternalServerError.Message,
		})
		_ = ctx.Render(
			"partials/x",
			bind,
			"partials/main",
		)
		return err
	}
	return nil
}
