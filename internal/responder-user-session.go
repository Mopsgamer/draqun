package internal

import (
	"errors"
	"fmt"
	"restapp/internal/environment"
	"restapp/internal/model"
	"restapp/internal/model_request"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/golang-jwt/jwt/v5"
)

// Uses the form request information.
func (r Responder) UserSignUp() error {
	id := "signup-error"
	req := new(model_request.UserSignUp)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.IsValidUserNick(req.Nickname) {
		return r.RenderWarning(MessageErrUserNick, id)
	}

	if !model.IsValidUserName(req.Username) {
		return r.RenderWarning(MessageErrUserName, id)
	}

	if user, _ := r.DB.UserByUsername(req.Username); user != nil {
		return r.RenderWarning(MessageErrUserExistsUsername, id)
	}

	if !model.IsValidUserPassword(req.Password) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if !model.IsValidUserEmail(req.Email) {
		return r.RenderWarning(MessageErrEmail, id)
	}

	if user, _ := r.DB.UserByEmail(req.Email); user != nil {
		return r.RenderWarning(MessageErrUserExistsEmail, id)
	}

	// TODO: phone validation
	// if !model.ValidatePhone(req.Phone) {
	// 	return r.RenderWarning(MessageErrPhone, id)
	// }

	// TODO: validate avatar and other properties

	if req.ConfirmPassword != req.Password {
		return r.RenderWarning(MessageErrBadConfirmPassword, id)
	}

	user, err := req.User()
	if err != nil {
		return r.RenderWarning(MessageFatalCanNotSignUp, id)
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
	req := new(model_request.UserLogin)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	if !model.IsValidUserPassword(req.Password) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if !model.IsValidUserEmail(req.Email) {
		return r.RenderWarning(MessageErrEmail, id)
	}

	user, err := r.DB.UserByEmail(req.Email)
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.Password) {
		return r.RenderWarning(MessageErrBadPassword, id)
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
	req := new(model_request.UserChangeName)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if req.NewNickname == user.Nick && req.NewUsername == user.Name {
		return r.RenderWarning(MessageErrUserNickSame, id)
	}

	if !model.IsValidUserNick(req.NewNickname) {
		return r.RenderWarning(MessageErrUserNick, id)
	}

	if !model.IsValidUserName(req.NewUsername) {
		return r.RenderWarning(MessageErrUserName, id)
	}

	if user, _ := r.DB.UserByUsername(req.NewUsername); user != nil {
		return r.RenderWarning(MessageErrUserExistsUsername, id)
	}

	user.Nick = req.NewNickname
	user.Name = req.NewUsername

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(MessageSuccChangedProfile, id)
}

// Uses the form request information.
func (r Responder) UserChangeEmail() error {
	id := "change-email-error"
	req := new(model_request.UserChangeEmail)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if req.NewEmail == user.Email {
		return r.RenderWarning(MessageErrEmailSame, id)
	}

	if !model.IsValidUserEmail(req.NewEmail) {
		return r.RenderWarning(MessageErrEmail, id)
	}

	if user, _ := r.DB.UserByEmail(req.NewEmail); user != nil {
		return r.RenderWarning(MessageErrUserExistsEmail, id)
	}

	if !model.IsValidUserPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrBadPassword, id)
	}

	user.Email = req.NewEmail

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(MessageSuccChangedEmail, id)
}

// Uses the form request information.
func (r Responder) UserChangePhone() error {
	id := "change-phone-error"
	req := new(model_request.UserChangePhone)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if req.NewPhone == user.Phone {
		return r.RenderWarning(MessageErrPhoneSame, id)
	}

	// TODO: Phone validation.
	// if !model.ValidatePhone(req.NewPhone) {
	// 	return r.RenderWarning(MessageErrPhone, id)
	// }

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrBadPassword, id)
	}

	user.Phone = req.NewPhone

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(MessageSuccChangedPhone, id)
}

// Uses the form request information.
func (r Responder) UserChangePassword() error {
	id := "change-password-error"
	req := new(model_request.UserChangePassword)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if req.NewPassword == user.Password {
		return r.RenderWarning(MessageErrPasswordSame, id)
	}

	if !model.IsValidUserPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrPassword, id)
	}

	if req.ConfirmPassword != req.NewPassword {
		return r.RenderWarning(MessageErrBadConfirmPassword, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrBadPassword, id)
	}

	user.Password = req.NewPassword

	err = r.DB.UserUpdate(*user)
	if err != nil {
		return err
	}

	r.HTMXRefresh()
	return r.RenderSuccess(MessageSuccChangedPass, id)
}

// Uses the form request information.
func (r Responder) UserDelete() error {
	id := "account-delete-error"
	req := new(model_request.UserDelete)
	err := r.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(MessageErrInvalidRequest, id)
	}

	user, err := r.GetOwner()
	if err != nil {
		return r.RenderWarning(MessageErrUserNotFound, id)
	}

	if user.Nick != req.ConfirmUsername {
		log.Warn(user.Nick)
		return r.RenderWarning(MessageErrBadUsernameConfirm, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(MessageErrBadPassword, id)
	}

	userOwnGroups, _ := r.DB.UserOwnGroups(user.Id)
	if len(userOwnGroups) > 0 {
		list := []string{}
		for groupIndex, group := range userOwnGroups {
			list[groupIndex] = group.Nick
		}
		return r.RenderDanger(MessageErrCanNotDeleteGroupOwnerAccount+" Groups: "+strings.Join(list, ", ")+".", id)
	}

	err = r.DB.UserDelete(user.Id)
	if err != nil {
		return err
	}

	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(MessageSuccDeletedUser, id)
}

// Authorize the user, using the current request information and new cookies.
func (r Responder) GiveToken(errorElementId string, user model.User) error {
	token, err := user.GenerateToken()
	if err != nil {
		log.Error(err)
		return r.RenderDanger(MessageFatalTokenGeneration, errorElementId)
	}

	r.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(model.UserTokenExpiration),
	})
	return r.RenderSuccess(MessageSuccLogin, errorElementId)
}

// Get the owner of the request using the "Authorization" header.
// Returns (nil, nil), if the header is empty.
func (r Responder) GetOwner() (*model.User, error) {
	authHeader := r.Cookies("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		err := errors.New("invalid authorization format. Expected Authorization header: Bearer and the token string")
		log.Error(err)
		return nil, err
	}

	tokenString := authHeader[7:]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			return nil, err
		}

		tokenBytes := []byte(environment.JWTKey)
		return tokenBytes, nil
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

	const prop = "Email"
	claims := token.Claims.(jwt.MapClaims)
	email, ok := claims[prop].(string)
	if !ok {
		err = fmt.Errorf("can not get claim property '"+prop+"'. Claims: %v", claims)
		return nil, err
	}

	user, err := r.DB.UserByEmail(email)
	return user, err
}
