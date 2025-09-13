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

func MapPage(ctx fiber.Ctx, db *model.DB, bind fiber.Map) fiber.Map {
	// empty values
	fiber.Locals(ctx, perms.LocalAuth, model.User{Db: db})
	fiber.Locals(ctx, perms.LocalGroup, model.Group{Db: db})
	fiber.Locals(ctx, perms.LocalMember, model.Member{Db: db})

	// actually fill
	_, err := perms.UserByAuthFromCtx(ctx, db)
	if err != nil && errors.Is(err, htmx.ErrToken) {
		log.Error(err)
	}
	_, _ = perms.GroupByIdFromCtx(ctx, db, "group_id")
	if fiber.Locals[model.Group](ctx, perms.LocalGroup).IsEmpty() {
		_, _ = perms.GroupByNameFromCtx(ctx, db, "group_name")
	}
	_ = perms.MemberByAuthAndGroupIdFromCtx(ctx, db, "group_id")

	// other values
	bindx := fiber.Map{
		"Ctx": ctx,
		"Db":  db,

		"AppName":    environment.AppName,
		"GitHubRepo": environment.GitHubRepo,
		"DenoJson":   environment.DenoJson,
		"GitJson":    environment.GitJson,
		"GoMod":      environment.GoMod,
	}

	maps.Copy(bindx, bind)
	return bindx
}
