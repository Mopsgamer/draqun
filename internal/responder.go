package internal

import (
	"fmt"
	"log"
	"regexp"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type Responder struct{ fiber.Ctx }

// Otherwise - json.
func (r Responder) IsHTMX() bool {
	return r.Ctx.Get("HX-Request", "") != ""
}

// Get 'partials/x' from '/partials/x?text=hello'.
func (r Responder) TemplateName() string {
	url := r.OriginalURL()[1:]
	url = regexp.MustCompilePOSIX("[a-zA-Z/-]+").FindString(url)
	return url
}

// This type describes ALL values in EVERY partial, which can be passed into ./templates/partials
// and used by htmx requests to replace DOM, using template generation through get requests
// EXAMPLE:
//
//	<div hx-get="/partials/settings">
//	<div hx-get="/partials/chat?class=compact">
//
// NOTE: wont move this to internal/htmx.go
// since its only for the RenderTemplate
type HTMXPartialQuery struct {
	Id      bool `query:"id"`
	Message bool `query:"id"`
	Open    bool `query:"open"`
}

func (r Responder) RenderTemplate() error {
	q := new(HTMXPartialQuery)
	err := r.Bind().Query(q)
	if err != nil {
		return err
	}
	return r.Render(r.TemplateName(), fiber.Map{
		"Id":      q.Id,
		"Message": q.Message,
		"Open":    q.Open,
	})
}

func (r Responder) RenderWarning(message, id string) error {
	return r.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

func (r Responder) UserRegister(db *Database) error {
	id := "auth-error"
	req := new(RegisterRequest)
	err := r.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBadPasswordMatch() {
		message := "Passwords do not match"
		return r.RenderWarning(message, id)
	}

	if req.IsMissing() {
		message := "Missing required fields"
		return r.RenderWarning(message, id)
	}

	user, err := req.CreateUser()
	if err != nil {
		log.Println(err)
		message := "Unable to register user"
		return r.RenderWarning(message, id)
	}

	err = db.UserSave(user)
	if err != nil {
		log.Println(err)
		message := "Unable to save user"
		return r.RenderWarning(message, id)
	}

	return r.Redirect().To("/")
}

func (r Responder) UserLogin(db *Database) error {
	id := "auth-error"
	req := new(LoginRequest)
	err := r.Ctx.Bind().JSON(req)
	log.Println("Request struct:", req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	// NOTE: Better compare passwords and then get the user?
	user, err := db.UserByEmail(req.Email)
	if err != nil {
		log.Println(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	token, err := user.GenerateJWT()
	if err != nil {
		message := "Error generating token"
		return r.RenderWarning(message, id)
	}

	return r.JSON(fiber.Map{
		"token": token,
	})
}

func (r Responder) GetUserData(db *Database) error {
	authHeader := r.Ctx.Get("Authorization")
	//log.Println(authHeader)
	if authHeader == "" {
		return r.JSON(fiber.Map{"error": "Authorization header missing"})
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return r.JSON(fiber.Map{"error": "Invalid authorization format"})
	}

	tokenString := authHeader[7:]

	//log.Println(tokenString)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return r.JSON(fiber.Map{"error": "Invalid token"})
	}

	claims := token.Claims.(jwt.MapClaims)

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return r.JSON(fiber.Map{"error": "User ID not found or is invalid"})
	}
	userID := int(userIDFloat)

	user, err := db.UserByID(userID)
	if err != nil {
		return r.JSON(fiber.Map{"error": "User not found"})
	}

	//log.Println(user)

	return r.JSON(user)
}
