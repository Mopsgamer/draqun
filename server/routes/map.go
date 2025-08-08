package routes

import (
	"maps"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
)

func MapPage(ctx fiber.Ctx, db *model.DB, bind fiber.Map) fiber.Map {
	// empty values
	fiber.Locals(ctx, perms.LocalAuth, model.User{Db: db})
	fiber.Locals(ctx, perms.LocalGroup, model.Group{Db: db})
	fiber.Locals(ctx, perms.LocalMember, model.Member{Db: db})

	// actually fill
	perms.UserByAuthFromCtx(ctx, db)
	perms.GroupByIdFromCtx(ctx, db, "group_id")
	if fiber.Locals[model.Group](ctx, perms.LocalGroup).IsEmpty() {
		perms.GroupByNameFromCtx(ctx, db, "group_name")
	}
	perms.MemberByAuthAndGroupIdFromCtx(ctx, db, "group_id")

	// other values
	bindx := fiber.Map{
		"Ctx": ctx,
		"Db":  db,

		"AppName":     environment.AppName,
		"GitHubRepo":  environment.GitHubRepo,
		"DenoJson":    environment.DenoJson,
		"GoMod":       environment.GoMod,
		"GitHash":     environment.GitHash,
		"GitHashLong": environment.GitHashLong,
	}

	maps.Copy(bindx, bind)
	return bindx
}
