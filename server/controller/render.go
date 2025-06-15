package controller

import (
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"

	"github.com/gofiber/fiber/v3"
)

func (ctl Controller) MapPage(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{
		"AppName":     environment.AppName,
		"GitHubRepo":  environment.GitHubRepo,
		"DenoJson":    environment.DenoJson,
		"GoMod":       environment.GoMod,
		"GitHash":     environment.GitHash,
		"GitHashLong": environment.GitHashLong,

		"User":   fiber.Locals[*model_database.User](ctl.Ctx, LocalAuth),
		"Group":  fiber.Locals[*model_database.Group](ctl.Ctx, LocalGroup),
		"Member": fiber.Locals[*model_database.Member](ctl.Ctx, LocalMember),
		"Rights": fiber.Locals[model_database.Role](ctl.Ctx, LocalRights),
	}

	bindx = MapMerge(&bindx, bind)
	return bindx
}

func (ctl Controller) RenderString(template string, bind any) (string, error) {
	return RenderString(ctl.Ctx.App(), template, bind)
}
