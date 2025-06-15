package controller

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type LocalKey int

const (
	LocalForm LocalKey = iota

	LocalAuth
	LocalGroup
	LocalMember
	LocalRights
)

func CheckGroup(groupId uint64) Handler {
	return func(ctl Controller) error {
		group := ctl.DB.GroupById(groupId)
		if group == nil {
			return environment.ErrGroupNotFound
		}

		ctl.Ctx.Locals(LocalGroup, group)
		return nil
	}
}

type RightsChecker func(role model_database.Role) bool

func CheckBindForm[T any](request T) Handler {
	return func(ctl Controller) error {
		if err := ctl.Ctx.Bind().Form(request); err != nil {
			return err
		}

		ctl.Ctx.Locals(LocalForm, request)
		return nil
	}
}

func CheckMember(must bool, groupId, userId uint64, rights RightsChecker) Handler {
	return func(ctl Controller) error {
		if err := CheckGroup(groupId)(ctl); err != nil {
			return err
		}

		member := ctl.DB.MemberById(groupId, userId)
		if !must {
			ctl.Ctx.Locals(LocalMember, member)
			return nil
		}

		if member == nil {
			return fmt.Errorf("group id = %d, member id = %d: not found", groupId, userId)
		}

		ctl.Ctx.Locals(LocalMember, member)

		role := ctl.DB.MemberRights(groupId, userId)
		if !rights(role) {
			return fiber.ErrForbidden
		}
		return nil
	}
}

func CheckAuthMember(must bool, groupId uint64, rights RightsChecker) Handler {
	return func(ctl Controller) error {
		if err := CheckAuth()(ctl); err != nil {
			return err
		}

		user := ctl.Ctx.Locals(LocalAuth).(*model_database.User)
		userId := user.Id
		return CheckMember(must, groupId, userId, rights)(ctl)
	}
}

func CheckAuth() Handler {
	return func(ctl Controller) error {
		var user *model_database.User
		authCookie := ctl.Ctx.Cookies("Authorization")
		if authCookie == "" {
			return fiber.ErrUnauthorized
		}

		if len(authCookie) < 8 || authCookie[:7] != "Bearer " {
			return fiber.ErrBadRequest
		}

		tokenString := authCookie[7:]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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

		user = ctl.DB.UserByEmail(email)
		if pass != user.Password {
			return errors.Join(environment.ErrToken, errors.New("incorrect password"))
		}

		ctl.Ctx.Locals("user", user)
		return nil
	}
}
