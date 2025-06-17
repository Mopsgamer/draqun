package controller

import (
	"io/fs"
	"log"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func NewChainFactory() func(handlers ...fiber.Handler) fiber.Handler {
	return func(handlers ...fiber.Handler) fiber.Handler {
		return func(ctx fiber.Ctx) error {
			for _, handler := range handlers {
				if err := handler(ctx); err != nil {
					return err
				}
			}
			return nil
		}
	}
}

func NewStaticFactory(embedFS fs.FS) func(dir string) fiber.Handler {
	return func(dir string) fiber.Handler {
		cacheDuration := time.Duration(0)
		if environment.BuildModeValue == environment.BuildModeProduction {
			cacheDuration = -1
		}
		if embedFS == nil {
			return static.New(dir, static.Config{Browse: true, CacheDuration: cacheDuration})
		}

		Fs, err := fs.Sub(embedFS, dir)
		if err != nil {
			log.Fatal(err)
		}

		return static.New("", static.Config{Browse: true, FS: Fs, CacheDuration: cacheDuration})
	}
}
