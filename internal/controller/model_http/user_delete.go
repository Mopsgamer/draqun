package model_http

import (
	"strings"
	"time"

	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

type UserDelete struct {
	CookieUserToken
	CurrentPassword string `form:"current-password"`
	ConfirmUsername string `form:"confirm-username"`
}

func (request *UserDelete) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "account-delete-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if user.Nick != request.ConfirmUsername {
		return ctl.RenderWarning(i18n.MessageErrBadUsernameConfirm, id)
	}

	if !user.CheckPassword(request.CurrentPassword) {
		return ctl.RenderWarning(i18n.MessageErrBadPassword, id)
	}

	userOwnGroups := ctl.DB.UserOwnGroupList(user.Id)
	if len(userOwnGroups) > 0 {
		list := []string{}
		for groupIndex, group := range userOwnGroups {
			list[groupIndex] = group.Nick
		}
		return ctl.RenderDanger(i18n.MessageErrCanNotDeleteGroupOwnerAccount+" Groups: "+strings.Join(list, ", ")+".", id)
	}

	if !ctl.DB.UserDelete(user.Id) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	ctl.HTMXRedirect(ctl.HTMXCurrentPath())
	return ctl.RenderSuccess(i18n.MessageSuccDeletedUser, id)
}
