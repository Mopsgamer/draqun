package restapp

import (
	"time"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req RegisterRequest) CreateUser() User {
	return User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  HashPassword(req.Password),
		CreatedAt: time.Now(),
	}
}
