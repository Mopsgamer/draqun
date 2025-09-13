package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/golang-jwt/jwt/v5"
)

func (user User) Claim() UserClaim {
	return UserClaim{
		Id:       user.Id,
		Password: user.Password,
	}
}

type UserClaim struct {
	Id       uint64
	Password PasswordHashed
}

func (c UserClaim) GetAudience() (jwt.ClaimStrings, error) {
	return []string{}, nil
}

func (c UserClaim) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(environment.UserAuthTokenExpiration)), nil
}

func (c UserClaim) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c UserClaim) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c UserClaim) GetIssuer() (string, error) {
	return "draqun", nil
}

func (c UserClaim) GetSubject() (string, error) {
	return "", nil
}
