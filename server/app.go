package internal

import (
	_ "embed"
	"io/fs"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS, clientEmbedded bool) (*fiber.App, error) {
	db, errDBLoad := model.InitDB()
	if errDBLoad != nil {
		log.Error(errDBLoad)
		return nil, errDBLoad
	}

	engine := NewAppHtmlEngine(db, embedFS, clientEmbedded, "client/templates")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			if htmx.IsHtmx(ctx) {
				return htmx.HandleHTMXError(ctx, err)
			}
			return err
		},
	})

	app.Use(logger.New())

	routes.RouteStatic(embedFS, clientEmbedded, app)
	routes.RoutePages(app, db)
	routes.RouteAccount(app, db)
	routes.RouteGroups(app, db)

	app.Use(func(ctx fiber.Ctx) error {
		if htmx.IsHtmx(ctx) {
			return htmx.TryRenderPage(
				ctx,
				"partials/alert",
				routes.MapPage(ctx, db, fiber.Map{
					"Variant": "primary",
					"Message": "404",
				}),
			)
		}
		if ctx.Method() == "GET" {
			return htmx.TryRenderPage(
				ctx,
				"partials/x",
				routes.MapPage(ctx, db, fiber.Map{
					"Title":         "404",
					"StatusCode":    fiber.StatusNotFound,
					"StatusMessage": fiber.ErrNotFound.Message,
					"CenterContent": true,
				}),
				"partials/main",
			)
		}

		return ctx.SendStatus(fiber.StatusNotFound)
	})

	return app, nil
}
