package model_http

import (
	"errors"
	"fmt"

	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
	"github.com/Mopsgamer/vibely/internal/environment"

	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

type CookieUserToken struct {
	UserToken string `cookie:"Authorization"`
}

// Get owner of the request using the "Authorization" header.
func (request *CookieUserToken) User(ctl controller_http.ControllerHttp) *model_database.User {
	if request.UserToken == "" {
		return nil
	}

	CatchTokenErr := func(err error) *model_database.User {
		log.Error(err)

		return nil
	}

	if len(request.UserToken) < 8 || request.UserToken[:7] != "Bearer " {
		err := errors.New("invalid authorization format. Expected: Bearer and the token string")
		return CatchTokenErr(err)
	}

	tokenString := request.UserToken[7:]

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

	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims["Email"].(string)
	if !ok {
		return CatchTokenErr(errors.New("expected any email from the token"))
	}

	pass, ok := claims["Password"].(string)
	if !ok {
		return CatchTokenErr(errors.New("expected any password from the token"))
	}

	user := ctl.DB.UserByEmail(email)
	if pass != user.Password {
		return CatchTokenErr(errors.New("got incorrect password from the token"))
	}

	return user
}
