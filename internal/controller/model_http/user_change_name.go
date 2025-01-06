package model_http

import (
	"restapp/internal/controller/controller_http"
	"restapp/internal/controller/model_database"
	"restapp/internal/i18n"
)

type UserChangeName struct {
	CookieUserToken
	NewNickname string `form:"new-nickname"`
	NewUsername string `form:"new-username"`
}

func (request *UserChangeName) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "change-name-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if request.NewNickname == user.Nick && request.NewUsername == user.Name {
		return ctl.RenderWarning(i18n.MessageErrUselessChange, id)
	}

	if !model_database.IsValidUserNick(request.NewNickname) {
		return ctl.RenderWarning(i18n.MessageErrUserNick, id)
	}

	if !model_database.IsValidUserName(request.NewUsername) {
		return ctl.RenderWarning(i18n.MessageErrUserName, id)
	}

	if ctl.DB.UserByUsername(request.NewUsername) != nil && request.NewNickname == user.Nick {
		return ctl.RenderWarning(i18n.MessageErrUserExistsUsername, id)
	}

	user.Nick = request.NewNickname
	user.Name = request.NewUsername

	if !ctl.DB.UserUpdate(*user) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRefresh()
	return ctl.RenderSuccess(i18n.MessageSuccChangedProfile, id)
}
