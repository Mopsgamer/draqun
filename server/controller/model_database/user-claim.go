package model_database

import (
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/golang-jwt/jwt/v5"
)

func (c User) GetAudience() (jwt.ClaimStrings, error) {
	return []string{c.Email}, nil
}

func (c User) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(environment.UserAuthTokenExpiration)), nil
}

func (c User) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c User) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c User) GetIssuer() (string, error) {
	return "draqun", nil
}

func (c User) GetSubject() (string, error) {
	return c.Email, nil
}
