package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// Uses the form request information.
func (r LogicHTTP) UserSignUp() error {
	id := "signup-error"
	req := new(model_request.UserSignUp)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	if !model_database.IsValidUserNick(req.Nickname) {
		return r.RenderWarning(i18n.MessageErrUserNick, id)
	}

	if !model_database.IsValidUserName(req.Username) {
		return r.RenderWarning(i18n.MessageErrUserName, id)
	}

	if r.DB.UserByUsername(req.Username) != nil {
		return r.RenderWarning(i18n.MessageErrUserExistsUsername, id)
	}

	if !model_database.IsValidUserPassword(req.Password) {
		return r.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !model_database.IsValidUserEmail(req.Email) {
		return r.RenderWarning(i18n.MessageErrEmail, id)
	}

	if r.DB.UserByEmail(req.Email) != nil {
		return r.RenderWarning(i18n.MessageErrUserExistsEmail, id)
	}

	// TODO: phone validation
	// if !model_database.ValidatePhone(req.Phone) {
	// 	return r.RenderWarning(i18n.MessageErrPhone, id)
	// }

	// TODO: validate avatar and other properties

	if req.ConfirmPassword != req.Password {
		return r.RenderWarning(i18n.MessageErrBadConfirmPassword, id)
	}

	user := req.User()
	if user == nil {
		return nil
	}

	if r.DB.UserCreate(*user) == nil {
		return r.RenderWarning(i18n.MessageFatalDatabaseQuery, id)
	}

	err = r.GiveToken(id, *user)
	if err != nil {
		return r.RenderWarning(i18n.MessageFatalTokenGeneration, id)
	}

	r.HTMXRedirect(r.HTMXCurrentPath())
	return nil
}

// Uses the form request information.
func (r LogicHTTP) UserLogin() error {
	id := "login-error"
	req := new(model_request.UserLogin)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	if !model_database.IsValidUserPassword(req.Password) {
		return r.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !model_database.IsValidUserEmail(req.Email) {
		return r.RenderWarning(i18n.MessageErrEmail, id)
	}

	user := r.DB.UserByEmail(req.Email)
	if user == nil {
		return r.RenderWarning(i18n.MessageErrUserNotFound, id)
	}

	if !user.CheckPassword(req.Password) {
		return r.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	err = r.GiveToken(id, *user)
	if err != nil {
		return r.RenderWarning(i18n.MessageFatalTokenGeneration, id)
	}

	r.HTMXRedirect(r.HTMXCurrentPath())
	return nil
}

// Sets the cookie through the request.
func (r LogicHTTP) UserLogout() error {
	r.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.Ctx.Render("partials/redirecting", fiber.Map{})
}

// Uses the form request information.
func (r LogicHTTP) UserChangeName() error {
	id := "change-name-error"
	req := new(model_request.UserChangeName)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if req.NewNickname == user.Nick && req.NewUsername == user.Name {
		return r.RenderWarning(i18n.MessageErrUserNickSame, id)
	}

	if !model_database.IsValidUserNick(req.NewNickname) {
		return r.RenderWarning(i18n.MessageErrUserNick, id)
	}

	if !model_database.IsValidUserName(req.NewUsername) {
		return r.RenderWarning(i18n.MessageErrUserName, id)
	}

	if r.DB.UserByUsername(req.NewUsername) != nil && req.NewNickname == user.Nick {
		return r.RenderWarning(i18n.MessageErrUserExistsUsername, id)
	}

	user.Nick = req.NewNickname
	user.Name = req.NewUsername

	if !r.DB.UserUpdate(*user) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return r.RenderSuccess(i18n.MessageSuccChangedProfile, id)
}

// Uses the form request information.
func (r LogicHTTP) UserChangeEmail() error {
	id := "change-email-error"
	req := new(model_request.UserChangeEmail)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if req.NewEmail == user.Email {
		return r.RenderWarning(i18n.MessageErrEmailSame, id)
	}

	if !model_database.IsValidUserEmail(req.NewEmail) {
		return r.RenderWarning(i18n.MessageErrEmail, id)
	}

	if r.DB.UserByEmail(req.NewEmail) != nil {
		return r.RenderWarning(i18n.MessageErrUserExistsEmail, id)
	}

	if !model_database.IsValidUserPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Email = req.NewEmail

	if !r.DB.UserUpdate(*user) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return r.RenderSuccess(i18n.MessageSuccChangedEmail, id)
}

// Uses the form request information.
func (r LogicHTTP) UserChangePhone() error {
	id := "change-phone-error"
	req := new(model_request.UserChangePhone)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if req.NewPhone == user.Phone {
		return r.RenderWarning(i18n.MessageErrPhoneSame, id)
	}

	// TODO: Phone validation.
	// if !model_database.ValidatePhone(req.NewPhone) {
	// 	return r.RenderWarning(i18n.MessageErrPhone, id)
	// }

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Phone = req.NewPhone

	if !r.DB.UserUpdate(*user) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return r.RenderSuccess(i18n.MessageSuccChangedPhone, id)
}

// Uses the form request information.
func (r LogicHTTP) UserChangePassword() error {
	id := "change-password-error"
	req := new(model_request.UserChangePassword)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if req.NewPassword == user.Password {
		return r.RenderWarning(i18n.MessageErrPasswordSame, id)
	}

	if !model_database.IsValidUserPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrPassword, id)
	}

	if req.ConfirmPassword != req.NewPassword {
		return r.RenderWarning(i18n.MessageErrBadConfirmPassword, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Password = req.NewPassword

	if !r.DB.UserUpdate(*user) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return r.RenderSuccess(i18n.MessageSuccChangedPass, id)
}

// Uses the form request information.
func (r LogicHTTP) UserDelete() error {
	id := "account-delete-error"
	req := new(model_request.UserDelete)
	err := r.Ctx.Bind().Form(req)
	if err != nil {
		return r.RenderWarning(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if user.Nick != req.ConfirmUsername {
		log.Warn(user.Nick)
		return r.RenderWarning(i18n.MessageErrBadUsernameConfirm, id)
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return r.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	userOwnGroups := r.DB.UserOwnGroupList(user.Id)
	if len(userOwnGroups) > 0 {
		list := []string{}
		for groupIndex, group := range userOwnGroups {
			list[groupIndex] = group.Nick
		}
		return r.RenderDanger(i18n.MessageErrCanNotDeleteGroupOwnerAccount+" Groups: "+strings.Join(list, ", ")+".", id)
	}

	if !r.DB.UserDelete(user.Id) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	r.HTMXRedirect(r.HTMXCurrentPath())
	return r.RenderSuccess(i18n.MessageSuccDeletedUser, id)
}

// Authorize the user, using the current request information and new cookies.
func (r LogicHTTP) GiveToken(errorElementId string, user model_database.User) error {
	token, err := user.GenerateToken()
	if err != nil {
		log.Error(err)
		return err
	}

	r.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(model_database.UserTokenExpiration),
	})

	return nil
}
