package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

var (
	messageFatalCannotRegister  = "Unable to register."
	messageFatalTokenGeneration = "Unable to create the token."
)

var (
	messageErrInvalidRequest = "Invalid request payload."
	messageErrBadPassConfirm = "Passwords are not same."
	messageErrBadNameConfirm = "Name are not same."
	messageErrBadPass        = "Invalid user password."
	messageErrUserNotFound   = "User not found."
)

var (
	messageSuccChangedProfile = "Successfully changed the user profile."
	messageSuccChangedPass    = "Successfully changed the user password."
	messageSuccChangedEmail   = "Successfully changed the user email."
	messageSuccChangedPhone   = "Successfully changed the user phone."
	messageSuccDeletedUser    = "Successfully deleted the user."
	messageSuccLogin          = "Successfully logged in! Redirecting..."
)

type Responder struct {
	fiber.Ctx
	DB Database
}

// Uses the form request information.
func (r Responder) UserRegister() error {
	id := "register-error"
	req := new(UserRegister)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	if req.ConfirmPassword != req.Password {
		return r.RenderWarning(messageErrBadPassConfirm, id)
	}

	user, err := req.CreateUser()
	if err != nil {
		return r.RenderWarning(messageFatalCannotRegister, id)
	}

	r.DB.UserCreate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.GiveToken(id, *user)
}

// Uses the form request information.
func (r Responder) UserLogin() error {
	id := "login-error"
	req := new(UserLogin)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	user, err := r.DB.UserByEmail(req.Email)
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.Password) {
		return r.RenderWarning(messageErrBadPass, id)
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
	return r.Render("partials/redirecting", fiber.Map{})
}

// Uses the form request information.
func (r Responder) UserChangeName() error {
	id := "change-name-error"
	req := new(UserChangeName)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	user.Name = req.NewName
	user.Tag = req.NewTag

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccChangedProfile, id)
}

// Uses the form request information.
func (r Responder) UserChangeEmail() error {
	id := "change-email-error"
	req := new(UserChangeEmail)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	user.Email = req.NewEmail

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccChangedEmail, id)
}

// Uses the form request information.
func (r Responder) UserChangePhone() error {
	id := "change-phone-error"
	req := new(UserChangePhone)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	user.Phone = req.NewPhone

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccChangedPhone, id)
}

// Uses the form request information.
func (r Responder) UserChangePassword() error {
	id := "change-password-error"
	req := new(UserChangePassword)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	if req.ConfirmPassword != req.NewPassword {
		return r.RenderWarning(messageErrBadPassConfirm, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	user.Password = req.NewPassword

	r.DB.UserUpdate(*user)

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccChangedPass, id)
}

// Uses the form request information.
func (r Responder) UserDelete() error {
	id := "account-delete-error"
	req := new(UserDelete)
	err := r.Bind().Form(req)
	if err != nil || req.IsBad() {
		return r.RenderWarning(messageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if user.Name != req.ConfirmName {
		log.Warn(user.Name)
		return r.RenderWarning(messageErrBadNameConfirm, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	r.DB.DeleteUser(*user)
	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccDeletedUser, id)
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user User) error {
	token, err := user.GenerateToken()
	if err != nil {
		return r.RenderWarning(messageFatalTokenGeneration, errorElementId)
	}

	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(tokenExpiration),
	})
	return r.RenderSuccess(messageSuccLogin, errorElementId)
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
		return nil, err
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, err
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		err = errors.New("invalid token")
		return nil, err
	}

	email := (token.Claims.(jwt.MapClaims))["email"].(string)

	user, err := r.DB.UserByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
