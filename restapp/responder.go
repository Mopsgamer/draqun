package restapp

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

type Responder struct{ Ctx fiber.Ctx }

func (c Responder) UserRegister(db *Database) error {
	req := new(RegisterRequest)
	if err := c.Ctx.Bind().JSON(req); err != nil {
		log.Println(err)
		return c.Ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		log.Println(req)
		return c.Ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	user := req.CreateUser()

	if err := db.UserSave(user); err != nil {
		log.Println(err)
		log.Println(req)
		return c.Ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to register user",
		})
	}

	return c.Ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}

func (c Responder) UserLogin(db *Database) error {
	req := new(LoginRequest)
	if err := c.Ctx.Bind().JSON(req); err != nil {
		//log.Println(err) // debug
		return c.Ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing email or password",
		})
	}

	user, err := db.UserByEmail(req.Email)
	if err != nil || !CheckPassword(user.Password, req.Password) {
		//log.Printf("Error: %s", err)
		return c.Ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	token, err := user.GenerateJWT()
	if err != nil {
		return c.Ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating token",
		})
	}

	return c.Ctx.JSON(fiber.Map{
		"token": token,
	})
}
