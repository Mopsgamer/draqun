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

func GroupByIdFromCtx(ctx fiber.Ctx, db *model.DB, groupIdUri string) (group model.Group, err error) {
	err = nil
	groupId := fiber.Params[uint64](ctx, groupIdUri)
	group, err = model.NewGroupFromId(db, groupId)
	if err != nil {
		err = htmx.ErrGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
}

func GroupByNameFromCtx(ctx fiber.Ctx, db *model.DB, groupNameUri string) (group model.Group, err error) {
	err = nil
	groupName := fiber.Params[string](ctx, groupNameUri)
	group, err = model.NewGroupFromName(db, groupName)
	if err != nil {
		err = htmx.ErrGroupNotFound
		return
	}

	ctx.Locals(LocalGroup, group)
	return
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
			return []byte{}, htmx.ErrToken.Join(err)
		}

		return []byte(environment.JWTKey), nil
	})

	if err != nil {
		err = htmx.ErrToken.Join(err)
	}
	return
}

func checkUser(db *model.DB, token *jwt.Token) (user model.User, err error) {
	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims["Email"].(string)
	if !ok || htmx.IsValidUserEmail(email) != nil {
		err = htmx.ErrToken.Join(errors.New("expected any email"))
		return
	}

	name, ok := claims["Name"].(string)
	if !ok || htmx.IsValidUserName(name) != nil {
		err = htmx.ErrToken.Join(errors.New("expected any name"))
		return
	}

	pass, ok := claims["Password"].(string)
	if !ok || htmx.IsValidUserPassword(pass) != nil {
		err = htmx.ErrToken.Join(errors.New("expected any password"))
		return
	}

	user, err = model.NewUserFromEmail(db, email)
	if err != nil {
		errName := user.FromName(name)
		if errName != nil {
			err = htmx.ErrUserNotFound.Join(err).Join(errName)
			return
		}
	}

	if pass != user.Password {
		err = htmx.ErrToken.Join(errors.New("incorrect password"))
	}
	return
}

func UserByAuthFromCtx(ctx fiber.Ctx, db *model.DB) (user model.User, err error) {
	user, err = model.User{Db: db}, nil
	tokenString := ctx.Cookies(fiber.HeaderAuthorization)

	token := new(jwt.Token)
	token, err = checkCookieToken(tokenString)
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
