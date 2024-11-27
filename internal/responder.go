package internal

import (
	"github.com/gofiber/fiber/v3"
)

type Responder struct {
	Ctx fiber.Ctx
	DB  Database
}

// Converts the pointer to the value
func (r Responder) MapMerge(maps ...*fiber.Map) fiber.Map {
	merge := fiber.Map{}
	for _, m := range maps {
		if m == nil {
			continue
		}

		for k, v := range *m {
			merge[k] = v
		}
	}

	return merge
}
