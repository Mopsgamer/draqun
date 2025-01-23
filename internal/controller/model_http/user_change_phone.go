package model_http

import (
	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
	"github.com/Mopsgamer/vibely/internal/i18n"
)

type UserChangePhone struct {
	CookieUserToken
	CurrentPassword string  `form:"current-password"`
	NewPhone        *string `form:"new-phone"`
}

func (request *UserChangePhone) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "change-phone-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if request.NewPhone == user.Phone {
		return ctl.RenderWarning(i18n.MessageErrPhoneSame, id)
	}

	if !model_database.IsValidUserPhone(request.NewPhone) {
		return ctl.RenderWarning(i18n.MessageErrPhone, id)
	}

	if !user.CheckPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	user.Phone = request.NewPhone

	if !ctl.DB.UserUpdate(*user) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return ctl.RenderSuccess(i18n.MessageSuccChangedPhone, id)
}
