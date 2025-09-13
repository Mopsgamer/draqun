package perms

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3"
)

const (
	LocalForm = "Form"

	LocalAuth   = "User"
	LocalGroup  = "Group"
	LocalMember = "Member"
	LocalRights = "Rights"
)

type RightsChecker func(ctx fiber.Ctx, role model.Role) bool

func GroupById(db *model.DB, groupIdUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		_, err := GroupByIdFromCtx(ctx, db, groupIdUri)
		if err != nil {
			return err
		}

		return ctx.Next()
	}
}

func GroupByName(db *model.DB, groupNameUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		_, err := GroupByIdFromCtx(ctx, db, groupNameUri)
		if err != nil {
			return err
		}

		return ctx.Next()
	}
}

func MemberByAuthAndGroupId(db *model.DB, groupIdUri string, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		err := MemberByAuthAndGroupIdFromCtx(ctx, db, groupIdUri)
		if err != nil {
			return err
		}

		member := fiber.Locals[model.Member](ctx, LocalRights)
		role := fiber.Locals[model.Role](ctx, LocalRights)
		if role.PermAdmin.Has() {
			return ctx.Next()
		}

		if member.IsAvailable() && (rights == nil || rights(ctx, role)) {
			return ctx.Next()
		}

		return htmx.AlertGroupMemberNotAllowed
	}
}

func UseBind[T any]() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := new(T)
		if err := ctx.Bind().All(request); err != nil {
			return err
		}

		ctx.Locals(LocalForm, *request)
		return ctx.Next()
	}
}

func UserByAuth(db *model.DB) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		_, err := UserByAuthFromCtx(ctx, db)
		if err != nil {
			return err
		}

		return ctx.Next()
	}
}
