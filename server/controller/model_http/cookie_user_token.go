package model_http

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/environment"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken      = errors.New("invalid authorization token format")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
)

type CookieUserToken struct {
	UserToken string `cookie:"Authorization"`
}

// Get owner of the request using the "Authorization" header.
func (request *CookieUserToken) User(ctl controller_http.ControllerHttp) (user *model_database.User, err error) {
	user = nil
	if request.UserToken == "" {
		return user, nil
	}

	if len(request.UserToken) < 8 || request.UserToken[:7] != "Bearer " {
		return user, ErrInvalidAuthHeader
	}

	tokenString := request.UserToken[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := errors.Join(ErrInvalidToken, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
			return user, err
		}

		tokenBytes := []byte(environment.JWTKey)
		return tokenBytes, nil
	})

	if err != nil {
		return user, errors.Join(ErrInvalidToken, err)
	}

	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims["Email"].(string)
	if !ok {
		return user, errors.Join(ErrInvalidToken, errors.New("expected any email"))
	}

	pass, ok := claims["Password"].(string)
	if !ok {
		return user, errors.Join(ErrInvalidToken, errors.New("expected any password"))
	}

	user = ctl.DB.UserByEmail(email)
	if pass != user.Password {
		return user, errors.Join(ErrInvalidToken, errors.New("incorrect password"))
	}

	return user, nil
}
