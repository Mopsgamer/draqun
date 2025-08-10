package model

import (
	"github.com/Mopsgamer/draqun/server/environment"

	"github.com/golang-jwt/jwt/v5"
)

// Get the token for the current user.
func (user User) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	return token.SignedString([]byte(environment.JWTKey))
}
