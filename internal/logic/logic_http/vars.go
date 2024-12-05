package logic_http

import (
	"errors"
	"fmt"
	"restapp/internal/environment"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"

	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

// Get owner of the request using the "Authorization" header.
// If the owner not found, returns (nil, nil), without errors.
// Automatically log-out and redirect to the home.
func (r *LogicHTTP) User() (user *model_database.User, err error) {
	authHeader := r.Ctx.Cookies("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	CatchTokenErr := func(err error) (*model_database.User, error) {
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
func (r *LogicHTTP) Group() *model_database.Group {
	groupUri := new(model_request.GroupUri)
	if err := r.Ctx.Bind().URI(groupUri); err != nil {
		log.Error(err)
		return nil
	}

	if groupUri.GroupId != 0 {
		group := r.DB.GroupById(groupUri.GroupId)
		return group
	}

	return nil
}
