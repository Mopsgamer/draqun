package model_http

import (
	"time"

	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
	"github.com/Mopsgamer/vibely/internal/i18n"

	"github.com/gofiber/fiber/v3/log"
)

type UserSignUp struct {
	*UserLogin
	Nickname        string  `form:"nickname"`
	Username        string  `form:"username"`
	Phone           *string `form:"phone"`
	ConfirmPassword string  `form:"confirm-password"`
}

func (request *UserSignUp) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "signup-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	if !model_database.IsValidUserNick(request.Nickname) {
		return ctl.RenderWarning(i18n.MessageErrUserNick, id)
	}

	if !model_database.IsValidUserName(request.Username) {
		return ctl.RenderWarning(i18n.MessageErrUserName, id)
	}

	if ctl.DB.UserByUsername(request.Username) != nil {
		return ctl.RenderWarning(i18n.MessageErrUserExistsUsername, id)
	}

	if !model_database.IsValidUserPassword(request.Password) {
		return ctl.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !model_database.IsValidUserEmail(request.Email) {
		return ctl.RenderWarning(i18n.MessageErrEmail, id)
	}

	if ctl.DB.UserByEmail(request.Email) != nil {
		return ctl.RenderWarning(i18n.MessageErrUserExistsEmail, id)
	}

	if !model_database.IsValidUserPhone(request.Phone) {
		return ctl.RenderWarning(i18n.MessageErrPhone, id)
	}

	// TODO: validate user avatar

	if request.ConfirmPassword != request.Password {
		return ctl.RenderWarning(i18n.MessageErrBadConfirmPassword, id)
	}

	hash, err := model_database.HashPassword(request.Password)
	if err != nil {
		log.Error(err)
		return nil
	}
	user := &model_database.User{
		Nick:      request.Nickname,
		Name:      request.Username,
		Email:     request.Email,
		Phone:     request.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
	}

	if ctl.DB.UserCreate(*user) == nil {
		return ctl.RenderWarning(i18n.MessageFatalDatabaseQuery, id)
	}

	err = request.GiveToken(ctl, *user)
	if err != nil {
		return ctl.RenderWarning(i18n.MessageFatalTokenGeneration, id)
	}

	ctl.HTMXRedirect(ctl.HTMXCurrentPath())
	return nil
}
