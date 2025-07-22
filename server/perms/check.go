package perms

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	LocalForm = "form"

	LocalAuth   = "user"
	LocalGroup  = "group"
	LocalMember = "member"
	LocalRights = "member-rights"
)

const AuthCookieKey = "Authorization"

type RightsChecker func(ctx fiber.Ctx, role database.Role) bool

func GroupById(db *goqu.Database, groupIdUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		groupFound, group := database.NewGroupFromId(db, groupId)
		if !groupFound {
			return htmx.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return ctx.Next()
	}
}

func GroupByName(db *goqu.Database, groupNameUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		groupName := fiber.Params[string](ctx, groupNameUri)
		groupFound, group := database.NewGroupFromName(db, groupName)
		if !groupFound {
			return htmx.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return ctx.Next()
	}
}

func MemberByAuthAndGroupId(db *goqu.Database, groupIdUri string, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := UserByAuth(db)(ctx); err != nil {
			return err
		}

		user := fiber.Locals[database.User](ctx, LocalAuth)
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		if err := GroupById(db, groupIdUri)(ctx); err != nil {
			return err
		}

		memberFound, member := database.NewMemberFromId(db, groupId, user.Id)
		if !memberFound { // never been a member
			return htmx.ErrGroupMemberNotFound
		}

		ctx.Locals(LocalMember, member)

		role := member.Role()
		if role.PermAdmin.Has() {
			return ctx.Next()
		}

		// TODO: implement ban and kick
		// should read action list
		isBannedOrKicked := false
		if !isBannedOrKicked && !rights(ctx, role) {
			return htmx.ErrGroupMemberNotAllowed
		}

		return ctx.Next()
	}
}

func UseForm[T any](request T) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := ctx.Bind().Form(request); err != nil {
			return err
		}

		ctx.Locals(LocalForm, request)
		return ctx.Next()
	}
}

func UserByAuth(db *goqu.Database) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		authCookie := ctx.Cookies(AuthCookieKey)
		if authCookie == "" {
			return fiber.ErrUnauthorized
		}

		if len(authCookie) < 1 {
			return htmx.ErrToken
		}

		token, err := jwt.Parse(authCookie, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				return []byte{}, errors.Join(htmx.ErrToken, err)
			}

			return []byte(environment.JWTKey), nil
		})

		if err != nil {
			return errors.Join(htmx.ErrToken, err)
		}

		claims := token.Claims.(jwt.MapClaims)
		email, ok := claims["Email"].(string)
		if !ok {
			return errors.Join(htmx.ErrToken, errors.New("expected any email"))
		}

		pass, ok := claims["Password"].(string)
		if !ok {
			return errors.Join(htmx.ErrToken, errors.New("expected any password"))
		}

		foundUser, user := database.NewUserFromEmail(db, email)
		if !foundUser {
			return htmx.ErrUserNotFound
		}

		if pass != user.Password {
			return errors.Join(htmx.ErrToken, errors.New("incorrect password"))
		}

		ctx.Locals(LocalAuth, user)
		return ctx.Next()
	}
}
