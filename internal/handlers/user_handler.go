package handlers

import (
	"time"

	"restapp/internal/models"
	"restapp/internal/services"

	"github.com/gofiber/fiber/v3"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func Register(c fiber.Ctx, us *services.UserService) error {
	req := new(RegisterRequest)
	if err := c.Bind().JSON(req); err != nil {
		//log.Println(err) // debug
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	user := models.User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  services.HashPassword(req.Password),
		CreatedAt: time.Now(),
	}

	if err := us.CreateUser(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c fiber.Ctx, us *services.UserService) error {
	req := new(LoginRequest)
	if err := c.Bind().JSON(req); err != nil {
		//log.Println(err) // debug
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing email or password",
		})
	}

	user, err := us.GetUserByEmail(req.Email)
	if err != nil || !services.CheckPassword(user.Password, req.Password) {
		//log.Printf("Error: %s", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	token, err := services.GenerateJWT(*user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}
