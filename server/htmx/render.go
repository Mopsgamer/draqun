package htmx

import (
	"flag"
	"maps"

	"github.com/gofiber/fiber/v3"
)

func TryRenderPage(ctx fiber.Ctx, name string, bind fiber.Map, layouts ...string) error {
	if len(layouts) > 0 && !IsHtmx(ctx) && CurrentHintBuilder != nil && flag.Lookup("test.v") == nil {
		if hints := CurrentHintBuilder(ctx, name, bind); len(hints) > 0 {
			_ = ctx.SendEarlyHints(hints)
		}
	}

	err := ctx.Render(name, bind, layouts...)
	maps.Copy(bind, fiber.Map{
		"Title":         "500",
		"StatusCode":    fiber.StatusInternalServerError,
		"StatusMessage": fiber.ErrInternalServerError.Message,
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
