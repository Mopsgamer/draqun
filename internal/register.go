package internal

import "time"

type RegisterRequest struct {
	Name            string `json:"name"`
	Tag             string `json:"tag"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm-password"`
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
