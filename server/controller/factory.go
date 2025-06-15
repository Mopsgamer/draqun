package controller

import (
	"github.com/gofiber/fiber/v3"
)

func NewChainFactory[H any](upgrade func(ctx fiber.Ctx) H) func(handlers ...func(ctl H) error) fiber.Handler {
	return func(handlers ...func(ctl H) error) fiber.Handler {
		return func(ctx fiber.Ctx) error {
			ctl := upgrade(ctx)
			for _, handler := range handlers {
				if err := handler(ctl); err != nil {
					return err
				}
			}
			return nil
		}
	}
}
