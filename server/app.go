package internal

import (
	_ "embed"
	"errors"
	"io/fs"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS, clientEmbedded bool) (*fiber.App, error) {
	engine := NewAppHtmlEngine(embedFS, clientEmbedded, "client/templates")
	app := fiber.New(fiber.Config{
		AppName:           environment.AppName,
		Views:             engine,
		PassLocalsToViews: true,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			if htmx.IsHtmx(ctx) {
				return htmx.HandleHTMXError(ctx, err)
			}
			if ctx.Method() == fiber.MethodGet {
				var fiberErr *fiber.Error
				if errors.As(err, &fiberErr) && fiberErr.Code == fiber.StatusNotFound {
					ctx.Status(fiber.StatusNotFound)
					return htmx.TryRenderPage(
						ctx,
						"partials/x",
						routes.MapPage(ctx, fiber.Map{
							"Title":         "404",
							"StatusCode":    fiber.StatusNotFound,
							"StatusMessage": fiber.ErrNotFound.Message,
						}),
						"partials/main",
					)
				}
			}
			return fiber.DefaultErrorHandler(ctx, err)
		},
	})

	app.Use(logger.New())

	routes.RouteHMR(app)
	routes.RouteStatic(embedFS, clientEmbedded, app)
	routes.RoutePages(app)
	routes.RouteAccount(app)
	routes.RouteGroups(app)
	routes.RouteNotifications(app)

	app.Hooks().OnPreStartupMessage(func(sm *fiber.PreStartupMessageData) error {
		sm.AddInfo("app-version", "Version", "\t\t"+versionString(clientEmbedded), 9)
		sm.AddInfo("app-version-git", "Git", "\t\t"+gitString(), 9)
		return nil
	})

	app.Use(func(ctx fiber.Ctx) error {
		if htmx.IsHtmx(ctx) {
			return htmx.TryRenderPage(
				ctx,
				"partials/alert",
				routes.MapPage(ctx, fiber.Map{
					"Variant": "primary",
					"Message": "404",
				}),
			)
		}
		if ctx.Method() == "GET" {
			ctx.Status(fiber.StatusNotFound)
			return htmx.TryRenderPage(
				ctx,
				"partials/x",
				routes.MapPage(ctx, fiber.Map{
					"Title":         "404",
					"StatusCode":    fiber.StatusNotFound,
					"StatusMessage": fiber.ErrNotFound.Message,
				}),
				"partials/main",
			)
		}

		return ctx.SendStatus(fiber.StatusNotFound)
	})

	return app, nil
}
