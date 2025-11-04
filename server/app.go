package internal

import (
	_ "embed"
	"fmt"
	"io/fs"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/routes"
	"github.com/fatih/color"

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
			return err
		},
	})

	app.Use(logger.New())

	routes.RouteStatic(embedFS, clientEmbedded, app)
	routes.RoutePages(app)
	routes.RouteAccount(app)
	routes.RouteGroups(app)

	clientEmbeddedColor := color.RGB(0, 180, 100)
	clientEmbeddedStatus := "client not embedded"
	if clientEmbedded {
		clientEmbeddedColor = color.New(color.FgHiRed)
		clientEmbeddedStatus = "client embedded"
	}
	hashColor := color.New(color.Faint)
	branchColor := color.RGB(100, 0, 180)
	version := branchColor.Sprint(link(environment.GitHubBranch, environment.GitJson.Branch)) + " " +
		"v" + environment.DenoJson.Version + " " +
		hashColor.Sprint(link(environment.GitHubCommit, environment.GitJson.Hash))
	app.Hooks().OnPreStartupMessage(func(sm *fiber.PreStartupMessageData) {
		sm.Header = ""
		scheme := "http"
		if sm.TLS {
			scheme = "https"
		}
		startedOn := ""
		if sm.Host == "0.0.0.0" {
			startedOn = fmt.Sprintf("\t%s%s://127.0.0.1:%s%s (bound on host 0.0.0.0 and port %s)",
				sm.ColorScheme.Blue, scheme, sm.Port, sm.ColorScheme.Reset, sm.Port)
		} else {
			startedOn = fmt.Sprintf("\t%s%s://%s:%s%s",
				sm.ColorScheme.Blue, scheme, sm.Host, sm.Port, sm.ColorScheme.Reset)
		}
		sm.PrimaryInfo = fiber.Map{
			"Application":          "\t" + environment.AppName,
			"Application version":  version,
			"Build mode":           "\t" + environment.BuildModeName,
			"Client embedded":      "\t" + clientEmbeddedColor.Sprint(clientEmbeddedStatus),
			"Server started on":    startedOn,
			"Total handlers count": sm.HandlerCount,
		}
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
