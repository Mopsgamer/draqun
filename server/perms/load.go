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

func GroupByIdFromCtx(ctx fiber.Ctx, groupIdUri string) (group model.Group, err error) {
	groupId := fiber.Params[uint64](ctx, groupIdUri)
	group, err = model.NewGroupFromId(groupId)
	if group.IsEmpty() {
		err = htmx.AlertGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
}

func GroupByNameFromCtx(ctx fiber.Ctx, groupNameUri string) (group model.Group, err error) {
	groupName := model.Name(fiber.Params[string](ctx, groupNameUri))
	group, err = model.NewGroupFromName(groupName)
	if group.IsEmpty() {
		err = htmx.AlertGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
}

func MemberByAuthAndGroupIdFromCtx(ctx fiber.Ctx, groupIdUri string) error {
	user, err := UserByAuthFromCtx(ctx)
	if err != nil {
		return err
	}

	groupId := fiber.Params[uint64](ctx, groupIdUri)
	_, err = GroupByIdFromCtx(ctx, groupIdUri)
	if err != nil {
		return err
	}

	member, err := model.NewMemberFromId(groupId, user.Id)
	if member.IsEmpty() { // never been a member
		return htmx.AlertGroupMemberNotFound
	}

	ctx.Locals(LocalMember, member)

	role := member.Role()
	ctx.Locals(LocalRights, role)
	return nil
}

func checkCookieToken(value string) (token *jwt.Token, err error) {
	if value == "" {
		err = htmx.AlertUserUnauthorized
		return
	}

	if len(value) < 1 {
		err = htmx.AlertToken
		return
	}

	token, err = jwt.Parse(value, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			return []byte{}, htmx.AlertToken.Join(err)
		}

		return []byte(environment.JWTKey), nil
	})

	if err != nil {
		err = htmx.AlertToken.Join(err)
	}
	return
}

func checkUser(token *jwt.Token) (user model.User, err error) {
	claims := token.Claims.(jwt.MapClaims)
	userIdFloat, ok := claims["Id"].(float64)
	userId := uint64(userIdFloat)
	if !ok || userId == 0 {
		err = htmx.AlertToken.Join(errors.New("expected non-zero id"))
		return
	}

	passwordString, ok := claims["Password"].(string)
	password := model.PasswordHashed(passwordString)
	if !ok || !password.IsValid() {
		err = htmx.AlertToken.Join(errors.New("expected any password"))
		return
	}

	user, err = model.NewUserFromId(userId)
	if user.IsEmpty() {
		err = htmx.AlertUserNotFound.Join(err)
		return
	}

	if password != user.Password {
		err = htmx.AlertToken.Join(errors.New("incorrect password"))
	}

	return
}

func UserByAuthFromCtx(ctx fiber.Ctx) (user model.User, err error) {
	user, err = model.User{}, nil
	tokenString := ctx.Cookies(fiber.HeaderAuthorization)

	token := new(jwt.Token)
	token, err = checkCookieToken(tokenString)
	if err != nil {
		return
	}

	user, err = checkUser(token)
	if err != nil {
		return
	}

	ctx.Locals(LocalAuth, user)
	return
}
