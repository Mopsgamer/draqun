package model_http

import (
	"restapp/internal/controller/controller_http"
	"restapp/internal/controller/model_database"
	"restapp/internal/i18n"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

type UserLogin struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (request *UserLogin) GiveToken(ctl controller_http.ControllerHttp, user model_database.User) error {
	token, err := user.GenerateToken()
	if err != nil {
		log.Error(err)
		return err
	}

	ctl.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + token,
		Expires: time.Now().Add(model_database.UserTokenExpiration),
	})

	return nil
}

func (request *UserLogin) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "login-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	if !model_database.IsValidUserPassword(request.Password) {
		return ctl.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !model_database.IsValidUserEmail(request.Email) {
		return ctl.RenderWarning(i18n.MessageErrEmail, id)
	}

	user := ctl.DB.UserByEmail(request.Email)
	if user == nil {
		return ctl.RenderWarning(i18n.MessageErrUserNotFound, id)
	}

	if !user.CheckPassword(request.Password) {
		return ctl.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	err := request.GiveToken(ctl, *user)
	if err != nil {
		return ctl.RenderWarning(i18n.MessageFatalTokenGeneration, id)
	}

	ctl.HTMXRedirect(ctl.HTMXCurrentPath())
	return nil
}
