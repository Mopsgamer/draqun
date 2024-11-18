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

// Converts user register request to the User struct.
func (req UserRegister) User() (*User, error) {
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	return &User{
		Nick:      req.Nickname,
		Name:      req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}, nil
}
