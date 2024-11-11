package model

import "time"

type UserRegister struct {
	Name            string `json:"name" form:"name"`
	Tag             string `json:"tag" form:"tag"`
	Email           string `json:"email" form:"email"`
	Phone           string `json:"phone" form:"phone"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm-password" form:"confirm-password"`
}

// Checks if the request contains invalid name, email or password fields.
func (req UserRegister) IsBad() bool {
	return req.Name == "" || req.Email == "" || req.Password == ""
}

func (req UserRegister) CreateUser() (*User, error) {
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	return &User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
	}, nil
}
