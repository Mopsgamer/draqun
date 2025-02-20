package model_http

import (
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/i18n"
)

type UserChangePassword struct {
	CookieUserToken
	CurrentPassword string `form:"current-password"`
	NewPassword     string `form:"new-password"`
	ConfirmPassword string `form:"confirm-password"`
}

func (request *UserChangePassword) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "change-password-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if request.NewPassword == user.Password {
		return ctl.RenderWarning(i18n.MessageErrPasswordSame, id)
	}

	if !model_database.IsValidUserPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrPassword, id)
	}

	if request.ConfirmPassword != request.NewPassword {
		return ctl.RenderWarning(i18n.MessageErrBadConfirmPassword, id)
	}

	if !user.CheckPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Password = request.NewPassword

	if !ctl.DB.UserUpdate(*user) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return ctl.RenderSuccess(i18n.MessageSuccChangedPass, id)
}
