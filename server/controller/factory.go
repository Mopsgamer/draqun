package controller

import (
	"github.com/gofiber/fiber/v3"
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
