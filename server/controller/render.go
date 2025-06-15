package controller

import (
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"

	"github.com/gofiber/fiber/v3"
)

func MapPage(ctx fiber.Ctx, bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{
		"AppName":     environment.AppName,
		"GitHubRepo":  environment.GitHubRepo,
		"DenoJson":    environment.DenoJson,
		"GoMod":       environment.GoMod,
		"GitHash":     environment.GitHash,
		"GitHashLong": environment.GitHashLong,

		"User":   fiber.Locals[*model_database.User](ctx, LocalAuth),
		"Group":  fiber.Locals[*model_database.Group](ctx, LocalGroup),
		"Member": fiber.Locals[*model_database.Member](ctx, LocalMember),
		"Rights": fiber.Locals[model_database.Role](ctx, LocalRights),
	}

	bindx = MapMerge(&bindx, bind)
	return bindx
}
