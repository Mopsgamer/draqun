package controller

import (
	"io/fs"
	"log"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type ChainHandler func(ctx fiber.Ctx) environment.HTMXAlert

func NewChainFactory() func(handlers ...fiber.Handler) fiber.Handler {
	return func(handlers ...fiber.Handler) fiber.Handler {
		return func(ctx fiber.Ctx) error {
			for _, handler := range handlers {
				if err := handler(ctx); err != nil {
					if IsHTMX(ctx) {
						return handleHtmxError(ctx, err)
					}
					return err
				}
			}
			return nil
		}
	}
}

func handleHtmxError(ctx fiber.Ctx, err error) error {
	level, message := environment.Danger, err.Error()
	if responseErr, ok := err.(environment.HTMXAlert); ok {
		level = responseErr.Level()
		message = responseErr.Local()
	}

	if ctx.Get("HX-Error-Wrap") == "false" {
		return ctx.SendString(message)
	}

	bind := fiber.Map{
		"Variant": level.String(),
		"Message": message,
	}
	buf, _ := RenderBuffer(ctx.App(), "partials/alert", bind)

	return ctx.Send(buf.Bytes())
}

func NewStaticFactory(embedFS fs.FS) func(dir string) fiber.Handler {
	return func(dir string) fiber.Handler {
		cacheDuration := time.Duration(-1)
		if environment.BuildModeValue == environment.BuildModeProduction {
			cacheDuration = time.Minute
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
