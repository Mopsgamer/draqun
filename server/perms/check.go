package perms

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	LocalForm = "Form"

	LocalAuth   = "User"
	LocalGroup  = "Group"
	LocalMember = "Member"
	LocalRights = "Rights"
)

type RightsChecker func(ctx fiber.Ctx, role model.Role) bool

func GroupByIdFromCtx(ctx fiber.Ctx, db *model.DB, groupIdUri string) (group model.Group, err error) {
	err = nil
	groupId := fiber.Params[uint64](ctx, groupIdUri)
	groupFound, group := model.NewGroupFromId(db, groupId)
	if !groupFound {
		err = htmx.ErrGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
}

func GroupByNameFromCtx(ctx fiber.Ctx, db *model.DB, groupNameUri string) (group model.Group, err error) {
	err = nil
	groupName := fiber.Params[string](ctx, groupNameUri)
	groupFound, group := model.NewGroupFromName(db, groupName)
	if !groupFound {
		err = htmx.ErrGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
}

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

func MemberByAuthAndGroupIdFromCtx(ctx fiber.Ctx, db *model.DB, groupIdUri string) error {
	user, err := UserByAuthFromCtx(ctx, db)
	if err != nil {
		return err
	}

	groupId := fiber.Params[uint64](ctx, groupIdUri)
	if err := GroupById(db, groupIdUri)(ctx); err != nil {
		return err
	}

	member, err := model.NewMemberFromId(db, groupId, user.Id)
	if err != nil { // never been a member
		return htmx.ErrGroupMemberNotFound
	}

	ctx.Locals(LocalMember, member)

	role := member.Role()
	ctx.Locals(LocalRights, role)
	return nil
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

		return htmx.ErrGroupMemberNotAllowed
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

func checkCookieToken(value string) (token *jwt.Token, err error) {
	if value == "" {
		err = fiber.ErrUnauthorized
		return
	}

	if len(value) < 1 {
		err = htmx.ErrToken
		return
	}

	token, err = jwt.Parse(value, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			return []byte{}, errors.Join(htmx.ErrToken, err)
		}

		return []byte(environment.JWTKey), nil
	})

	err = errors.Join(htmx.ErrToken, err)
	return
}

func checkUser(db *model.DB, token *jwt.Token) (user model.User, err error) {
	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims["Email"].(string)
	if !ok {
		err = errors.Join(htmx.ErrToken, errors.New("expected any email"))
		return
	}

	pass, ok := claims["Password"].(string)
	if !ok {
		err = errors.Join(htmx.ErrToken, errors.New("expected any password"))
		return
	}

	user, err = model.NewUserFromEmail(db, email)
	if err != nil {
		err = htmx.ErrUserNotFound
		return
	}

	if pass != user.Password {
		err = errors.Join(htmx.ErrToken, errors.New("incorrect password"))
		return
	}

	return
}

func UserByAuthFromCtx(ctx fiber.Ctx, db *model.DB) (user model.User, err error) {
	user, err = model.User{Db: db}, nil
	cookieToken := ctx.Cookies(fiber.HeaderAuthorization)

	token, err := checkCookieToken(cookieToken)
	if err != nil {
		return
	}

	user, err = checkUser(db, token)
	if err != nil {
		return
	}

	ctx.Locals(LocalAuth, user)
	return
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
