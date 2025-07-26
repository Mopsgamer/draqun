package controller

import (
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/perms"
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

		"User":   ctx.Locals(perms.LocalAuth),
		"Group":  ctx.Locals(perms.LocalGroup),
		"Member": ctx.Locals(perms.LocalMember),
		"Rights": ctx.Locals(perms.LocalRights),
	}

	bindx = MapMerge(&bindx, bind)
	return bindx
}

// Converts the pointer to a value
func MapMerge(maps ...*fiber.Map) fiber.Map {
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
