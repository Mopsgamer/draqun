package internal

import (
	"errors"
	"fmt"
	"restapp/internal/model"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

var (
	MessageFatalCannotRegister  = "Unable to register."
	MessageFatalTokenGeneration = "Unable to create the token."
)

var (
	MessageErrInvalidRequest     = "Invalid request payload."
	MessageErrPassword           = "Invalid password pattern. " + model.MessageDetailPassword
	MessageErrNickname           = "Invalid nickname pattern. " + model.MessageDetailNickname
	MessageErrUsername           = "Invalid username pattern. " + model.MessageDetailUsername
	MessageErrEmail              = "Invalid email pattern. " + model.MessageDetailEmail
	MessageErrPhone              = "Invalid phone number pattern. " + model.MessageDetailPhone
	messageErrBadPassConfirm     = "Passwords are not same."
	messageErrBadUsernameConfirm = "Usernames are not same."
	messageErrBadPass            = "Invalid user password."
	messageErrUserNotFound       = "User not found."
	messageErrUserExistsUsername = "This username is taken."
	messageErrUserExistsEmail    = "This email is taken."
	// messageErrUserExistsPhone    = "This phone number is taken."
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
	req := new(model.UserRegister)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.ValidateNickname(req.Nickname) {
		return r.RenderWarning(MessageErrNickname, id)
	}

	if !model.ValidateUsername(req.Username) {
		return r.RenderWarning(MessageErrUsername, id)
	}

	if user, _ := r.DB.UserByUsername(req.Username); user != nil {
		return r.RenderWarning(messageErrUserExistsUsername, id)
	}

	if !model.ValidatePassword(req.Password) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if !model.ValidateEmail(req.Email) {
		return r.RenderWarning(MessageErrEmail, id)
	}

	if user, _ := r.DB.UserByEmail(req.Email); user != nil {
		return r.RenderWarning(messageErrUserExistsEmail, id)
	}

	// TODO: Phone validation.
	// if !model.ValidatePhone(req.Phone) {
	// 	return r.RenderWarning(MessageErrPhone, id)
	// }

	if req.ConfirmPassword != req.Password {
		return r.RenderWarning(messageErrBadPassConfirm, id)
	}

	user, err := req.User()
	if err != nil {
		return r.RenderWarning(MessageFatalCannotRegister, id)
	}

	err = r.DB.UserCreate(*user)
	if err != nil {
		return err
	}

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.GiveToken(id, *user)
}

// Uses the form request information.
func (r Responder) UserLogin() error {
	id := "login-error"
	req := new(model.UserLogin)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.ValidatePassword(req.Password) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if !model.ValidateEmail(req.Email) {
		return r.RenderWarning(MessageErrEmail, id)
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
	req := new(model.UserChangeName)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.ValidateNickname(req.NewNickname) {
		return r.RenderWarning(MessageErrNickname, id)
	}

	if !model.ValidateUsername(req.NewUsername) {
		return r.RenderWarning(MessageErrUsername, id)
	}

	if user, _ := r.DB.UserByUsername(req.NewUsername); user != nil {
		return r.RenderWarning(messageErrUserExistsUsername, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	user.Nickname = req.NewNickname
	user.Username = req.NewUsername

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(messageSuccChangedProfile, id)
}

// Uses the form request information.
func (r Responder) UserChangeEmail() error {
	id := "change-email-error"
	req := new(model.UserChangeEmail)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.ValidateEmail(req.NewEmail) {
		return r.RenderWarning(MessageErrEmail, id)
	}

	if user, _ := r.DB.UserByEmail(req.NewEmail); user != nil {
		return r.RenderWarning(messageErrUserExistsEmail, id)
	}

	if !model.ValidatePassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	user.Email = req.NewEmail

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(messageSuccChangedEmail, id)
}

// Uses the form request information.
func (r Responder) UserChangePhone() error {
	id := "change-phone-error"
	req := new(model.UserChangePhone)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	// TODO: Phone validation.
	// if !model.ValidatePhone(req.NewPhone) {
	// 	return r.RenderWarning(MessageErrPhone, id)
	// }

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	user.Phone = req.NewPhone

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(messageSuccChangedPhone, id)
}

// Uses the form request information.
func (r Responder) UserChangePassword() error {
	id := "change-password-error"
	req := new(model.UserChangePassword)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.ValidatePassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrPassword, id)
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

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(messageSuccChangedPass, id)
}

// Uses the form request information.
func (r Responder) UserDelete() error {
	id := "account-delete-error"
	req := new(model.UserDelete)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(messageErrUserNotFound, id)
	}

	if user.Nickname != req.ConfirmUsername {
		log.Warn(user.Nickname)
		return r.RenderWarning(messageErrBadUsernameConfirm, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(messageErrBadPass, id)
	}

	err = r.DB.DeleteUser(user.Id)
	if err != nil {
		return err
	}

	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(messageSuccDeletedUser, id)
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user model.User) error {
	token, err := user.GenerateToken()
	if err != nil {
		return r.RenderWarning(MessageFatalTokenGeneration, errorElementId)
	}

	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(model.UserTokenExpiration),
	})
	return r.RenderSuccess(messageSuccLogin, errorElementId)
}

// Get the owner of the request using the "Authorization" header.
// Returns (nil, nil), if the header is empty.
func (r Responder) GetOwner() (*model.User, error) {
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
		return model.JwtKey, nil
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
	return user, err
}
