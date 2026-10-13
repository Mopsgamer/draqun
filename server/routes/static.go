package routes

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func RouteStatic(embedFS fs.FS, clientEmbedded bool, app *fiber.App) {
	cfg := static.Config{
		Browse:        false,
		CacheDuration: time.Minute,
	}
	if environment.BuildEnvironment == environment.BuildModeDevelopment {
		cfg.CacheDuration = -1
		cfg.ModifyResponse = func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return nil
		}
		app.Use("/static", func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return c.Next()
		})
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

	_ = os.MkdirAll("uploads", 0755)
	app.Get("/uploads/:filename", func(c fiber.Ctx) error {
		filename := filepath.Base(c.Params("filename"))
		filePath := filepath.Join("uploads", filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("File not found")
		}
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; media-src 'self'; style-src 'unsafe-inline'")
		c.Set("X-Frame-Options", "DENY")
		c.Type(filepath.Ext(filename))
		return c.Send(content)
	})
}
