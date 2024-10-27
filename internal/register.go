package internal

import "time"

type RegisterRequest struct {
	Name            string `form:"name"`
	Tag             string `form:"tag"`
	Email           string `form:"email"`
	Phone           string `form:"phone"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm-password"`
}

func (req RegisterRequest) IsBadPasswordMatch() bool {
	return req.Password != req.ConfirmPassword
}

func (req RegisterRequest) IsMissing() bool {
	return req.Name == "" || req.Email == "" || req.Password == ""
}

func (req RegisterRequest) CreateUser() (User, error) {
	hash, err := HashPassword(req.Password)
	return User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
	}, err
}
