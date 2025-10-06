package perms

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

const (
	LocalForm = "Form"

	LocalAuth   = "User"
	LocalGroup  = "Group"
	LocalMember = "Member"
	LocalRights = "Rights"
)

type RightsChecker func(ctx fiber.Ctx, role model.Role) bool

func GroupById(groupIdUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		group, err := GroupByIdFromCtx(ctx, groupIdUri)
		if err != nil {
			return err
		}

		if !group.IsAvailable() {
			return htmx.AlertGroupNotFound
		}

		return ctx.Next()
	}
}

func GroupByName(groupNameUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		group, err := GroupByNameFromCtx(ctx, groupNameUri)
		if err != nil {
			return err
		}

		if !group.IsAvailable() {
			return htmx.AlertGroupNotFound
		}

		return ctx.Next()
	}
}

func MemberByAuthAndGroupId(groupIdUri string, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		err := MemberByAuthAndGroupIdFromCtx(ctx, groupIdUri)
		if err != nil {
			return err
		}

		member := fiber.Locals[model.Member](ctx, LocalMember)
		if !member.IsAvailable() {
			return htmx.AlertGroupMemberNotFound
		}

		group := fiber.Locals[model.Group](ctx, LocalGroup)
		role := fiber.Locals[model.Role](ctx, LocalRights)
		if role.PermAdmin.Has() || member.UserId == group.OwnerId {
			return ctx.Next()
		}

		if rights == nil || !rights(ctx, role) {
			return htmx.AlertGroupMemberNotAllowed
		}

		return ctx.Next()
	}
}

func UseBind[T any]() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		request := new(T)
		if err := ctx.Bind().All(request); err != nil {
			return err
		}

		log.Debug(string(ctx.BodyRaw()))
		ctx.Locals(LocalForm, *request)
		return ctx.Next()
	}
}

func UserByAuth() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		_, err := UserByAuthFromCtx(ctx)
		if err != nil {
			return err
		}

		return ctx.Next()
	}
}
