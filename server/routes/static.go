package routes

import (
	"io/fs"
	"log"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func RouteStatic(embedFS fs.FS, clientEmbedded bool, app *fiber.App) {
	app.Get("/static*", staticHandler(embedFS, clientEmbedded, environment.StaticFolder))
}

func staticHandler(embedFS fs.FS, clientEmbedded bool, dir string) fiber.Handler {
	cacheDuration := time.Duration(-1)
	if environment.BuildEnvironment == environment.BuildModeProduction {
		cacheDuration = time.Minute
	}
	cfg := static.Config{
		Browse:        true,
		CacheDuration: cacheDuration,
		// NotFoundHandler: func(ctx fiber.Ctx) error { return ctx.Next() },
	}
	if !clientEmbedded {
		return static.New(dir, cfg)
	}

	Fs, err := fs.Sub(embedFS, dir)
	if err != nil {
		log.Fatal(err)
	}

	cfg.FS = Fs
	return static.New(dir, cfg)
}
