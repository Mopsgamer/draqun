package controller

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model_database"
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

type RightsChecker func(ctx fiber.Ctx, role model_database.Role) bool

func PopulatePage(db *database.Database) fiber.Handler {
	groupIdUri := "group_id"
	groupNameUri := "group_name"
	return func(ctx fiber.Ctx) error {
		_ = CheckAuth(db)(ctx)
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		groupName := fiber.Params[string](ctx, groupNameUri)
		var group *model_database.Group
		if groupName != "" {
			group = db.GroupByName(groupName)
		} else {
			group = db.GroupById(groupId)
		}
		if group != nil {
			_ = CheckAuthMember(db, groupIdUri, nil)(ctx)
		}
		return ctx.Next()
	}
}

func CheckGroup(db *database.Database, groupIdUri string) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		group := db.GroupById(groupId)
		if group == nil {
			return environment.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return ctx.Next()
	}
}

func CheckAuthMember(db *database.Database, groupIdUri string, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := CheckAuth(db)(ctx); err != nil {
			return err
		}

		user := fiber.Locals[*model_database.User](ctx, LocalAuth)
		groupId := fiber.Params[uint64](ctx, groupIdUri)
		if err := CheckGroup(db, groupIdUri)(ctx); err != nil {
			return err
		}

		member := db.MemberById(groupId, user.Id)

		if member == nil {
			return environment.ErrGroupMemberNotFound
		}

		ctx.Locals(LocalMember, member)

		role := db.MemberRights(groupId, user.Id)
		if rights != nil && !rights(ctx, role) {
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

func CheckAuth(db *database.Database) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var user *model_database.User
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

		user = db.UserByEmail(email)
		if pass != user.Password {
			return errors.Join(environment.ErrToken, errors.New("incorrect password"))
		}

		ctx.Locals(LocalAuth, user)
		return ctx.Next()
	}
}
