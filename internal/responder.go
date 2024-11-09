package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

type Responder struct {
	fiber.Ctx
	DB Database
}

// Uses the current request information.
func (r Responder) UserRegister() error {
	id := "auth-error"
	req := new(RegisterRequest)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
		message := "Unable to register user"
		return r.RenderWarning(message, id)
	}

	return r.GiveToken(id, *user)
}

// Uses the current request information.
func (r Responder) UserLogin() error {
	id := "auth-error"
	req := new(LoginRequest)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.DB.UserByEmail(req.Email)
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.Password) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.GiveToken(id, *user)
}

// Uses the current request information.
func (r Responder) UserLogout() error {
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.Render("partials/auth-logout", fiber.Map{})
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user User) error {
	token, err := user.GenerateToken()
	if err != nil {
		message := "Error generating token"
		return r.RenderWarning(message, errorElementId)
	}

	message := "Success! Redirecting..."
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(tokenExpiration),
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

	userIDFloat, ok := claims["id"].(float64)
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
