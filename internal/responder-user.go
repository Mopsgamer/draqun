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

// Uses the form request information.
func (r Responder) UserRegister() error {
	id := "auth-error"
	req := new(UserRegister)
	err := r.Bind().Form(req)
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

	r.DB.UserCreate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.GiveToken(id, *user)
}

// Uses the form request information.
func (r Responder) UserLogin() error {
	id := "auth-error"
	req := new(UserLogin)
	err := r.Bind().Form(req)
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

// Sets the cookie through the request.
func (r Responder) UserLogout() error {
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.Render("partials/auth-logout", fiber.Map{})
}

// Uses the form request information.
func (r Responder) UserChangeName() error {
	id := "change-error"
	req := new(UserChangeName)
	err := r.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	user.Name = req.NewName
	user.Tag = req.NewTag

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess("Profile has been changed successfully.", id)
}

// Uses the form request information.
func (r Responder) UserChangeEmail() error {
	id := "change-error"
	req := new(UserChangeEmail)
	err := r.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.CurrentPassword) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	user.Email = req.NewEmail

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess("Email has been changed successfully.", id)
}

// Uses the form request information.
func (r Responder) UserChangePhone() error {
	id := "change-error"
	req := new(UserChangePhone)
	err := r.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.CurrentPassword) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	user.Phone = req.NewPhone

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess("Phone number has been changed successfully.", id)
}

// Uses the form request information.
func (r Responder) UserChangePassword() error {
	id := "change-error"
	req := new(UserChangePassword)
	err := r.Bind().Form(req)
	if err != nil {
		log.Error(err)
		message := "Invalid request payload"
		return r.RenderWarning(message, id)
	}

	if req.IsBad() {
		message := "Missing email or password"
		return r.RenderWarning(message, id)
	}

	if req.IsBadPasswordMatch() {
		message := "Passwords do not match"
		return r.RenderWarning(message, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		log.Error(err)
		message := "User not found"
		return r.RenderWarning(message, id)
	}

	if !CheckPassword(user.Password, req.CurrentPassword) {
		message := "Invalid email or password"
		return r.RenderWarning(message, id)
	}

	user.Password = req.NewPassword

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess("Email has been changed successfully.", id)
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user User) error {
	token, err := user.GenerateToken()
	if err != nil {
		log.Error(err)
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
		err := errors.New("invalid authorization format. Expected Authorization header: Bearer and the token string")
		log.Error(err)
		return nil, err
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			log.Error(err)
			return nil, err
		}
		return secretKey, nil
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	if !token.Valid {
		err = errors.New("invalid token")
		log.Error(err)
		return nil, err
	}

	email := (token.Claims.(jwt.MapClaims))["email"].(string)

	user, err := r.DB.UserByEmail(email)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return user, nil
}
