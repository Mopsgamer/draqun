package controller

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"
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

type RightsChecker func(role model_database.Role) bool

func PopulateGroup(db *database.Database, groupId uint64) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		CheckGroup(db, groupId)(ctx)
		return nil
	}
}

func PopulateMember(db *database.Database, groupId, userId uint64, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		CheckMember(db, groupId, userId, rights)(ctx)
		return nil
	}
}

func PopulateAuthMember(db *database.Database, groupId uint64, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		CheckAuthMember(db, groupId, rights)(ctx)
		return nil
	}
}

func PopulateAuth(db *database.Database) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		CheckAuth(db)(ctx)
		return nil
	}
}

func CheckGroup(db *database.Database, groupId uint64) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		group := db.GroupById(groupId)
		if group == nil {
			return environment.ErrGroupNotFound
		}

		ctx.Locals(LocalGroup, group)
		return nil
	}
}

func CheckMember(db *database.Database, groupId, userId uint64, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := CheckGroup(db, groupId)(ctx); err != nil {
			return err
		}

		member := db.MemberById(groupId, userId)

		if member == nil {
			return fmt.Errorf("group id = %d, member id = %d: not found", groupId, userId)
		}

		ctx.Locals(LocalMember, member)

		role := db.MemberRights(groupId, userId)
		if !rights(role) {
			return fiber.ErrForbidden
		}
		return nil
	}
}

func CheckAuthMember(db *database.Database, groupId uint64, rights RightsChecker) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := CheckAuth(db)(ctx); err != nil {
			return err
		}

		user := ctx.Locals(LocalAuth).(*model_database.User)
		userId := user.Id
		return CheckMember(db, groupId, userId, rights)(ctx)
	}
}

func CheckBindForm[T any](request T) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		if err := ctx.Bind().Form(request); err != nil {
			return err
		}

		ctx.Locals(LocalForm, request)
		return nil
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
		return nil
	}
}
