package model_database

import (
	"github.com/Mopsgamer/vibely/internal/environment"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Check the encoded password with the user's normal password.
// The normal password should not be encoded.
func (user User) CheckPassword(password string) bool {
	return CheckPassword(user.Password, password)
}

// Get the token for the current user.
func (user User) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)
	return token.SignedString([]byte(environment.JWTKey))
}

// Encode the normal password string.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// Check the encoded password with the user's normal password.
// The normal password should not be encoded.
func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
