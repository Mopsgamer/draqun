package htmx

import (
	"github.com/gofiber/fiber/v3"
)

func TryRenderPage(ctx fiber.Ctx, name string, bind any, layouts ...string) error {
	err := ctx.Render(name, bind, layouts...)
	if err != nil {
		ctx.Render(
			"partials/x",
			fiber.Map{
				"Title":         "500",
				"StatusCode":    fiber.StatusInternalServerError,
				"StatusMessage": fiber.ErrInternalServerError.Message,
				"CenterContent": true,
			},
			"partials/main",
		)
		return err
	}
	return nil
}
