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
	fiber.Locals(ctx, perms.LocalGroup, model.Group{})
	fiber.Locals(ctx, perms.LocalMember, model.Member{})

	// actually fill
	_, err := perms.UserByAuthFromCtx(ctx)
	if err != nil && errors.Is(err, htmx.ErrToken) {
		log.Error(err)
	}
	_, _ = perms.GroupByIdFromCtx(ctx, "group_id")
	if fiber.Locals[model.Group](ctx, perms.LocalGroup).IsEmpty() {
		_, _ = perms.GroupByNameFromCtx(ctx, "group_name")
	}
	_ = perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id")

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
