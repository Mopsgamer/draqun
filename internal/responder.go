package internal

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type Responder struct {
	fiber.Ctx
	DB Database
}

// Otherwise graphql or something.
func (r Responder) IsHTMX() bool {
	return r.Ctx.Get("HX-Request", "") != ""
}

// Get 'partials/x' from '/partials/x?text=hello'.
func (r Responder) TemplateName() string {
	url := r.OriginalURL()[1:]
	url = regexp.MustCompilePOSIX("[a-zA-Z/-]+").FindString(url)
	return url
}

func (r Responder) RenderPage(templatePath, title string, layouts ...string) error {
	user, err := r.GetOwner()
	m := fiber.Map{
		"Title":      "Restapp - " + title,
		"User":       user,
		"TokenError": err != nil,
		"GoodUser":   user != nil,
		"Message":    "Authorization error",
		"Id":         "local-token-error",
	}
	return r.Render(templatePath, m, layouts...)
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
	Id      string `query:"id"`
	Message string `query:"id"`
	Open    bool   `query:"open"`
	Token   string `query:"token"` // its safe
	User    User   `query:"user"`  // its safe
}

func (r Responder) RenderTemplate() error {
	q := new(HTMXPartialQuery)
	err := r.Bind().Query(q)
	r.GetOwner()
	if err != nil {
		return err
	}
	return r.Render(r.TemplateName(), fiber.Map{
		"Id":      q.Id,
		"Message": q.Message,
		"Open":    q.Open,
		"Token":   q.Token, // its safe
		"User":    q.User,  // its safe
	})
}

func (r Responder) RenderDanger(message, id string) error {
	return r.Render("partials/danger", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

func (r Responder) RenderWarning(message, id string) error {
	return r.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

func (r Responder) RenderSuccess(message, id string) error {
	return r.Render("partials/success", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

func (r Responder) UserRegister() error {
	id := "auth-error"
	req := new(RegisterRequest)
	err := r.Ctx.Bind().Form(req)
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

	return r.GiveToken(id, *user)
}

func (r Responder) UserLogin() error {
	id := "auth-error"
	req := new(LoginRequest)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		log.Println(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.DB.UserByEmail(req.Email)
	if err != nil {
		log.Println(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	return r.GiveToken(id, *user)
}

func (r Responder) UserLogout() error {
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	return r.Render("partials/auth-logout", fiber.Map{})
}

func (r Responder) GiveToken(errorElementId string, user User) error {
	token, err := user.GenerateJWT()
	if err != nil {
		message := "Error generating token"
		return r.RenderWarning(message, errorElementId)
	}

	message := "Success! Redirecting..."
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(30 * 24 * time.Hour), // 30 days
	})
	return r.Render("partials/auth-success", fiber.Map{
		"Id":      errorElementId,
		"Message": message,
		"Token":   token,
	})
}

// Get the owner of the request using the "Authorization" header.
// Returns (nil, nil), if the header is empty.
func (r Responder) GetOwner() (*User, error) {
	authHeader := r.Cookies("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("invalid authorization format. Expected Authorization header: Bearer and the token string")
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("user ID not found or is invalid")
	}
	userID := int(userIDFloat)

	user, err := r.DB.UserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
