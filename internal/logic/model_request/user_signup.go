package model_request

import (
	"restapp/internal/logic/model_database"
	"time"

	"github.com/gofiber/fiber/v3/log"
)

type UserSignUp struct {
	Nickname        string  `form:"nickname"`
	Username        string  `form:"username"`
	Email           string  `form:"email"`
	Phone           *string `form:"phone"`
	Password        string  `form:"password"`
	ConfirmPassword string  `form:"confirm-password"`
}

// Converts user sign up request to the User struct.
func (req UserSignUp) User() *model_database.User {
	hash, err := model_database.HashPassword(req.Password)
	if err != nil {
		log.Error(err)
		return nil
	}
	return &model_database.User{
		Nick:      req.Nickname,
		Name:      req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}
}
