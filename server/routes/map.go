package routes

import (
	"maps"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func MapPage(ctx fiber.Ctx, bind fiber.Map) fiber.Map {
	bindx := fiber.Map{
		"AppName":     environment.AppName,
		"GitHubRepo":  environment.GitHubRepo,
		"DenoJson":    environment.DenoJson,
		"GoMod":       environment.GoMod,
		"GitHash":     environment.GitHash,
		"GitHashLong": environment.GitHashLong,

		"User":   ctx.Locals(perms.LocalAuth),
		"Group":  ctx.Locals(perms.LocalGroup),
		"Member": ctx.Locals(perms.LocalMember),
		"Rights": ctx.Locals(perms.LocalRights),
	}

	maps.Copy(bindx, bind)
	return bindx
}
