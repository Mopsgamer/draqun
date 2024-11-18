package model

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// The server's secret key.
var JwtKey = []byte(os.Getenv("JWT_KEY"))

// User token expiration: 24 Hours.
var UserTokenExpiration = 24 * time.Hour

// The user as a database entry
type User struct {
	Id       uint   `db:"id"`
	Nickname string `db:"nickname"`
	Username string `db:"username"`
	Email    string `db:"email"`
	Phone    string `db:"phone"`
	// Hashed password string
	Password  string    `db:"password"`
	Avatar    string    `db:"avatar"`
	CreatedAt time.Time `db:"created_at"`
	LastSeen  time.Time `db:"last_seen"`
}

func (c User) GetAudience() (jwt.ClaimStrings, error) {
	return []string{c.Email}, nil
}

func (c User) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(UserTokenExpiration)), nil
}

func (c User) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c User) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (c User) GetIssuer() (string, error) {
	return "restapp", nil
}

func (c User) GetSubject() (string, error) {
	return c.Email, nil
}

// Check the encoded password with the current user struct password.
func (user User) CheckPassword(password string) bool {
	return CheckPassword(user.Password, password)
}

// Get the token for the current user.
func (user User) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)

	return token.SignedString(JwtKey)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
