package internal

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/routes"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/mattn/go-colorable"
)

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS, clientEmbedded bool) (*fiber.App, error) {
	htmx.OnRenderError = func(name string, err error, path string) {
		model.RecordOtelSpan(
			"TEMPLATE ERROR: "+name,
			"TEMPLATE RENDER",
			fiber.StatusInternalServerError,
			0,
			map[string]string{
				"template.name":  name,
				"template.error": err.Error(),
				"http.path":      path,
			},
		)
	}
	engine := NewAppHtmlEngine(embedFS, clientEmbedded, "client/templates")
	app := fiber.New(fiber.Config{
		AppName:           environment.AppName,
		Views:             engine,
		PassLocalsToViews: true,
		ColorScheme:       fiber.DefaultColors,
		ErrorHandler: func(ctx fiber.Ctx, err error) error {
			if err != nil {
				model.RecordOtelSpan(
					"SERVER ERROR: "+ctx.Path(),
					"ERROR HANDLER",
					fiber.StatusInternalServerError,
					0,
					map[string]string{
						"error":     err.Error(),
						"http.path": ctx.Path(),
					},
				)
			}

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

	app.Use(logger.New(logger.Config{
		ForceColors: true,
	}))

	app.Use(func(ctx fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()
		duration := time.Since(start)

		status := ctx.Response().StatusCode()
		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		model.RecordOtelSpan(
			ctx.Method()+" "+ctx.Path(),
			"HTTP SERVER",
			status,
			duration,
			map[string]string{
				"http.method":      ctx.Method(),
				"http.path":        ctx.Path(),
				"http.status_code": fmt.Sprintf("%d", status),
			},
		)

		return err
	})

	routes.RouteHMR(app)
	routes.RouteStatic(embedFS, clientEmbedded, app)
	routes.RoutePages(app)
	routes.RouteAccount(app)
	routes.RouteGroups(app)
	routes.RouteNotifications(app)
	routes.RouteUsers(app)
	routes.RouteFriends(app)

	app.Hooks().OnPreStartupMessage(func(sm *fiber.PreStartupMessageData) error {
		sm.AddInfo("app-version", "Version", "\t\t"+versionString(clientEmbedded), 9)
		sm.AddInfo("app-version-git", "Git", "\t\t"+gitString(), 9)

		if os.Getenv("FORCE_COLOR") == "1" || os.Getenv("CLICOLOR_FORCE") == "1" {
			sm.PreventDefault = true
			out := colorable.NewColorableStdout()
			colors := fiber.DefaultColors

			fmt.Fprintf(out, "%s\n", fmt.Sprintf("\n    _______ __\n   / ____(_) /_  ___  _____\n  / /_  / / __ \\/ _ \\/ ___/\n / __/ / / /_/ /  __/ /\n/_/   /_/_.___/\\___/_/          %s", colors.Red+"v"+fiber.Version+colors.Reset))
			fmt.Fprintln(out, strings.Repeat("-", 50))

			fmt.Fprintf(out, "%sINFO%s Server started on: \t%shttp://127.0.0.1:%s (bound on host 0.0.0.0 and port %s)%s\n", colors.Green, colors.Reset, colors.Blue, environment.Port, environment.Port, colors.Reset)
			fmt.Fprintf(out, "%sINFO%s Application name: \t%s%s%s\n", colors.Green, colors.Reset, colors.Blue, environment.AppName, colors.Reset)
			fmt.Fprintf(out, "%sINFO%s Version: \t\t%s%s%s\n", colors.Green, colors.Reset, colors.Blue, versionString(clientEmbedded), colors.Reset)
			fmt.Fprintf(out, "%sINFO%s Git: \t\t\t%s%s%s\n", colors.Green, colors.Reset, colors.Blue, gitString(), colors.Reset)
			fmt.Fprintf(out, "%sINFO%s PID: \t\t\t%s%d%s\n", colors.Green, colors.Reset, colors.Blue, os.Getpid(), colors.Reset)
			fmt.Fprintln(out)
		}
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
