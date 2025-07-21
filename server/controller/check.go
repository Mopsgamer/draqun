package controller

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
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

func CheckGroupById(db *goqu.Database, groupIdUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		group := database.Group{Db: db}
		if !group.FromId(groupId) {
			return environment.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return ctx.Next()
	}
}

func CheckGroupByName(db *goqu.Database, groupNameUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		groupName := fiber.Params[string](ctx, groupNameUri)
		group := database.Group{Db: db}
		if !group.FromName(groupName) {
			return environment.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return ctx.Next()
	}
}

func CheckAuthMember(db *goqu.Database, groupIdUri string, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := CheckAuth(db)(ctx); err != nil {
			return err
		}

		user := fiber.Locals[database.User](ctx, LocalAuth)
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		if err := CheckGroupById(db, groupIdUri)(ctx); err != nil {
			return err
		}

		member := database.Member{Db: db}
		if !member.FromId(groupId, user.Id) { // never been a member
			return environment.ErrGroupMemberNotFound
		}

		ctx.Locals(LocalMember, member)

		role := member.Role()
		if role.PermAdmin.Has() {
			return ctx.Next()
		}

		// TODO: implement ban and kick
		// should read action list for both from db
		isBannedOrKicked := false
		if !isBannedOrKicked && !rights(ctx, role) {
			return environment.ErrGroupMemberNotAllowed
		}

		return ctx.Next()
	}
}

func CheckBindForm[T any](request T) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := ctx.Bind().Form(request); err != nil {
			return err
		}

		ctx.Locals(LocalForm, request)
		return ctx.Next()
	}
}

func CheckAuth(db *goqu.Database) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		user := database.User{Db: db}
		authCookie := ctx.Cookies(AuthCookieKey)
		if authCookie == "" {
			return fiber.ErrUnauthorized
		}

		if len(authCookie) < 1 {
			return environment.ErrToken
		}

		token, err := jwt.Parse(authCookie, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				return user, errors.Join(environment.ErrToken, err)
			}

			tokenBytes := []byte(environment.JWTKey)
			return tokenBytes, nil
		})

		if err != nil {
			return errors.Join(environment.ErrToken, err)
		}

		claims := token.Claims.(jwt.MapClaims)
		email, ok := claims["Email"].(string)
		if !ok {
			return errors.Join(environment.ErrToken, errors.New("expected any email"))
		}

		pass, ok := claims["Password"].(string)
		if !ok {
			return errors.Join(environment.ErrToken, errors.New("expected any password"))
		}

		if !user.FromEmail(email) {
			return environment.ErrUserNotFound
		}

		if pass != user.Password {
			return errors.Join(environment.ErrToken, errors.New("incorrect password"))
		}

		ctx.Locals(LocalAuth, user)
		return ctx.Next()
	}
}
