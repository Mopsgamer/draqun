package internal

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// The server's secret key.
var secretKey = []byte(os.Getenv("JWT_KEY"))

// User token expiration: 24 Hours.
var tokenExpiration = 24 * time.Hour

// The user token validator.
var tokenValidator = jwt.NewValidator()

// The user as a json or
type User struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Tag       string    `json:"tag" db:"tag"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Password  string    `json:"password" db:"password"`
	Avatar    string    `json:"avatar" db:"avatar"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (c User) GetAudience() (jwt.ClaimStrings, error) {
	return []string{c.Email}, nil
}

func (c User) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(tokenExpiration)), nil
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

// Encode the password.
func (user User) HashPassword() (string, error) {
	return HashPassword(user.Password)
}

// Check the encoded password with the current user struct password.
func (user User) CheckPassword(hashedPassword string) bool {
	return CheckPassword(hashedPassword, user.Password)
}

// Get the token for the current user.
func (user User) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, user)

	return token.SignedString(secretKey)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
