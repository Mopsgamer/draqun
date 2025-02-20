package model_http

import (
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/i18n"
)

type UserChangeEmail struct {
	CookieUserToken
	CurrentPassword string `form:"current-password"`
	NewEmail        string `form:"new-email"`
}

func (request *UserChangeEmail) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "change-email-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if request.NewEmail == user.Email {
		return ctl.RenderWarning(i18n.MessageErrEmailSame, id)
	}

	if !model_database.IsValidUserEmail(request.NewEmail) {
		return ctl.RenderWarning(i18n.MessageErrEmail, id)
	}

	if ctl.DB.UserByEmail(request.NewEmail) != nil {
		return ctl.RenderWarning(i18n.MessageErrUserExistsEmail, id)
	}

	if !model_database.IsValidUserPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrPassword, id)
	}

	if !user.CheckPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Email = request.NewEmail

	if !ctl.DB.UserUpdate(*user) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return ctl.RenderSuccess(i18n.MessageSuccChangedEmail, id)
}
