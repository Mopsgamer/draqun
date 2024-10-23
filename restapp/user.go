package restapp

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID        uint      `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Tag       string    `json:"tag" db:"tag"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Password  string    `json:"-" db:"password"`
	Avatar    string    `json:"avatar" db:"avatar"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (user User) GenerateJWT() (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "restapp",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func (user User) Save(db *sqlx.DB) error {
	query := `INSERT INTO users (name, tag, email, phone, password, avatar, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt)
	if err != nil {
		log.Println(err)
		log.Println(user)
		return err
	}
	return nil
}
