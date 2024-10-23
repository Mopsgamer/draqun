package restapp

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"
)

func UserRegister(db *sqlx.DB, c fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.Bind().JSON(req); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	user := User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  HashPassword(req.Password),
		CreatedAt: time.Now(),
	}

	if err := user.Save(db); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

func UserLogin(db *sqlx.DB, c fiber.Ctx) error {
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

	user, err := UserByEmail(db, req.Email)
	if err != nil || !CheckPassword(user.Password, req.Password) {
		//log.Printf("Error: %s", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	token, err := user.GenerateJWT()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func UserByEmail(db *sqlx.DB, email string) (*User, error) {
	var user = new(User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
              FROM users WHERE email = ?`
	err := db.Get(&user, query, email)
	if err != nil {
		log.Println(err)
		log.Println(user)
		return nil, errors.New("user not found")
	}
	return user, nil
}
