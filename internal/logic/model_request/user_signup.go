package model_request

import (
	"restapp/internal/logic/model"
	"time"

	"github.com/gofiber/fiber/v3/log"
)

type UserSignUp struct {
	Nickname        string `json:"nickname" form:"nickname"`
	Username        string `json:"username" form:"username"`
	Email           string `json:"email" form:"email"`
	Phone           string `json:"phone" form:"phone"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm-password" form:"confirm-password"`
}

// Converts user sign up request to the User struct.
func (req UserSignUp) User() *model.User {
	hash, err := model.HashPassword(req.Password)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &model.User{
		Nick:      req.Nickname,
		Name:      req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
}
