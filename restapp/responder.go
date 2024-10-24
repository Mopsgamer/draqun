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

func (c Responder) RenderWarning(message, id string) error {
	log.Println("rendering...")
	return c.Ctx.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

func (c Responder) UserRegister(db *Database) error {
	id := "auth-error"
	req := new(RegisterRequest)
	err := c.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return c.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing required fields"
		return c.RenderWarning(message, id)
	}

	user, err := req.CreateUser()
	if err != nil {
		log.Println(err)
		message := "Unable to register user"
		return c.RenderWarning(message, id)
	}

	err = db.UserSave(user)
	if err != nil {
		log.Println(err)
		message := "Unable to save user"
		return c.RenderWarning(message, id)
	}

	return c.Ctx.Redirect().To("/")
}

func (c Responder) UserLogin(db *Database) error {
	id := "auth-error"
	req := new(LoginRequest)
	err := c.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return c.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return c.RenderWarning(message, id)
	}

	user, err := db.UserByEmail(req.Email)
	if err != nil {
		log.Println(err)
		message := "User not found"
		return c.RenderWarning(message, id)
	}
	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return c.RenderWarning(message, id)
	}

	token, err := user.GenerateJWT()
	if err != nil {
		message := "Error generating token"
		return c.RenderWarning(message, id)
	}

	// the client should save the token...
	_ = token
	return c.Ctx.Redirect().To("/")
}
