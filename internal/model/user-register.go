package model

import "time"

type UserRegister struct {
	Nickname        string `json:"nickname" form:"nickname"`
	Username        string `json:"username" form:"username"`
	Email           string `json:"email" form:"email"`
	Phone           string `json:"phone" form:"phone"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm-password" form:"confirm-password"`
}

func (req UserRegister) CreateUser() (*User, error) {
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	return &User{
		Nickname:  req.Nickname,
		Username:  req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}, nil
}
