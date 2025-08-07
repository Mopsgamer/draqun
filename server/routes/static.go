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
	app.Get("/static", NewStaticFactory(embedFS, clientEmbedded)(environment.StaticFolder))
}

func NewStaticFactory(embedFS fs.FS, clientEmbedded bool) func(dir string) fiber.Handler {
	return func(dir string) fiber.Handler {
		cacheDuration := time.Duration(-1)
		if environment.BuildModeValue == environment.BuildModeProduction {
			cacheDuration = time.Minute
		}
		if !clientEmbedded {
			return static.New(dir, static.Config{Browse: true, CacheDuration: cacheDuration})
		}

		Fs, err := fs.Sub(embedFS, dir)
		if err != nil {
			log.Fatal(err)
		}

		return static.New("", static.Config{Browse: true, FS: Fs, CacheDuration: cacheDuration})
	}
}
