package logic

import (
	"bytes"
	"errors"
	"fmt"
	"restapp/internal/environment"
	"restapp/internal/logic/database"
	"restapp/internal/logic/model"
	"restapp/internal/logic/model_request"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

type LogicCtx interface {
	Cookies(key string, defaultValue ...string) string
	Bind() *fiber.Bind
	App() *fiber.App
}

type Logic struct {
	Ctx LogicCtx
	DB  *database.Database
}

func (Logic) UserLogout() error {
	return nil
}

// Converts the pointer to the value
func MapMerge(maps ...*fiber.Map) fiber.Map {
	merge := fiber.Map{}
	for _, m := range maps {
		if m == nil {
			continue
		}

		for k, v := range *m {
			merge[k] = v
		}
	}

	return merge
}

func (r *Logic) RenderBuffer(template string, bind any) (bytes.Buffer, error) {
	buf := bytes.NewBuffer([]byte{})
	err := r.Ctx.App().Config().Views.Render(buf, template, bind)

	return *buf, err
}

// Get owner of the request using the "Authorization" header.
// If the owner not found, returns (nil, nil), without errors.
// Automatically log-out and redirect to the home.
func (r *Logic) User() (user *model.User, err error) {
	authHeader := r.Ctx.Cookies("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	CatchTokenErr := func(err error) (*model.User, error) {
		log.Error(err)
		err = r.UserLogout()
		if err != nil {
			log.Error(err)
		}

		return nil, err
	}

	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		err := errors.New("invalid authorization format. Expected Authorization header: Bearer and the token string")
		return CatchTokenErr(err)
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, err
		}

		tokenBytes := []byte(environment.JWTKey)
		return tokenBytes, nil
	})

	if err != nil {
		return CatchTokenErr(err)
	}

	if !token.Valid {
		err = errors.New("invalid token")
		return CatchTokenErr(err)
	}

	const prop = "Email"
	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims[prop].(string)
	if !ok {
		err = fmt.Errorf("can not get claim property '"+prop+"'. Claims: %v", claims)
		return CatchTokenErr(err)
	}

	return r.DB.UserByEmail(email), nil
}

// Get group by the id from current URI.
func (r *Logic) Group() *model.Group {
	// FIXME: can't get group from websocket uri
	groupUri := new(model_request.GroupUri)
	if err := r.Ctx.Bind().URI(groupUri); err != nil {
		log.Error(err)
		return nil
	}

	if groupUri.GroupId != nil {
		group := r.DB.GroupById(*groupUri.GroupId)
		return group
	}

	return nil
}
