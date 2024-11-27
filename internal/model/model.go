package model

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

func Map(value any) *fiber.Map {
	var result *fiber.Map
	bytes, _ := json.Marshal(value)
	json.Unmarshal(bytes, result)
	return result
}
