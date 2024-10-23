package restapp

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

type Responder struct{ Ctx fiber.Ctx }

// Otherwise - json.
func (c Responder) IsHTMX() bool {
	return c.Ctx.Get("HX-Request", "") != ""
}

func (c Responder) UserRegister(db *Database) error {
	req := new(RegisterRequest)
	err := c.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return c.Ctx.SendString(message)
	}

	if req.IsBad() {
		message := "Missing required fields"
		return c.Ctx.SendString(message)
	}

	user, err := req.CreateUser()
	if err != nil {
		log.Println(err)
		message := "Unable to register user"
		return c.Ctx.SendString(message)
	}

	err = db.UserSave(user)
	if err != nil {
		log.Println(err)
		message := "Unable to save user"
		return c.Ctx.SendString(message)
	}

	return c.Ctx.Redirect().To("/")
}

func (c Responder) UserLogin(db *Database) error {
	req := new(LoginRequest)
	err := c.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return c.Ctx.SendString(message)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return c.Ctx.SendString(message)
	}

	user, err := db.UserByEmail(req.Email)
	if err != nil {
		log.Println(err)
		message := "User not found"
		return c.Ctx.SendString(message)
	}
	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return c.Ctx.SendString(message)
	}

	token, err := user.GenerateJWT()
	if err != nil {
		message := "Error generating token"
		return c.Ctx.SendString(message)
	}

	// the client should save the token...
	_ = token
	return c.Ctx.Redirect().To("/")
}
