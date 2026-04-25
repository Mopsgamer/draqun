package routes

import (
	"errors"
	"maps"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func MapPage(ctx fiber.Ctx, bind fiber.Map) fiber.Map {
	// empty values
	fiber.Locals(ctx, perms.LocalAuth, model.User{})
	// actually fill
	_, err := perms.UserByAuthFromCtx(ctx)
	if err != nil && errors.Is(err, htmx.ErrToken) {
		log.Error(err)
	}

	// other values
	bindx := fiber.Map{
		"Ctx": ctx,

		"AppName":      environment.AppName,
		"GitHubRepo":   environment.GitHubRepo,
		"GitHubCommit": environment.GitHubCommit,
		"GitHubBranch": environment.GitHubBranch,
		"DenoJson":     environment.DenoJson,
		"GitJson":      environment.GitJson,
		"GoMod":        environment.GoMod,
	}

	maps.Copy(bindx, bind)
	return bindx
}
