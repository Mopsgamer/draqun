package routes

import (
	"io/fs"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func RouteStatic(embedFS fs.FS, clientEmbedded bool, app *fiber.App) {
	cfg := static.Config{
		Browse:        false,
		CacheDuration: time.Minute,
	}

	if !clientEmbedded {
		app.Use("/static", static.New("dist/static", cfg))
	} else {
		subFS, err := fs.Sub(embedFS, "dist/static")
		if err != nil {
			log.Fatal(err)
		}
		cfg.FS = subFS
		app.Use("/static", static.New("", cfg))
	}
}
