package model_http

import (
	"fmt"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/Mopsgamer/draqun/server/i18n"
)

const (
	GroupCreateModePublic  = model_database.GroupModePublic
	GroupCreateModePrivate = model_database.GroupModePrivate
)

type GroupCreate struct {
	CookieUserToken
	Name        string  `form:"name"`
	Nick        string  `form:"nick"`
	Password    *string `form:"password"`
	Mode        int     `form:"mode"`
	Description string  `form:"description"`
	Avatar      string  `form:"avatar"`
}

func (request *GroupCreate) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "new-group-error"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	user, _ := request.User(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if ctl.DB.GroupByName(request.Name) != nil {
		return ctl.RenderDanger(i18n.MessageErrGroupExistsGroupname, id)
	}

	if !model_database.IsValidGroupName(request.Name) {
		return ctl.RenderDanger(i18n.MessageErrGroupName, id)
	}

	if !model_database.IsValidGroupNick(request.Nick) {
		return ctl.RenderDanger(i18n.MessageErrGroupNick, id)
	}

	if !model_database.IsValidGroupPassword(request.Password) {
		return ctl.RenderDanger(i18n.MessageErrGroupPassword, id)
	}

	if !model_database.IsValidGroupDescription(request.Description) {
		return ctl.RenderDanger(i18n.MessageErrGroupDescription, id)
	}

	if !model_database.IsValidGroupMode(request.Mode) {
		return ctl.RenderDanger(fmt.Sprintf(i18n.MessageErrGroupMode+" Got: '%d'.", request.Mode), id)
	}

	// TODO: validate group avatar

	group := &model_database.Group{
		CreatorId:   user.Id,
		Nick:        request.Nick,
		Name:        request.Name,
		Mode:        model_database.GroupMode(request.Mode),
		Description: request.Description,
		Password:    request.Password,
		Avatar:      request.Avatar,
		CreatedAt:   time.Now(),
	}
	groupId := ctl.DB.GroupCreate(*group)
	if groupId == nil {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	group.Id = *groupId

	member := model_database.Member{
		GroupId:  group.Id,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  true,
		IsBanned: false,
	}

	if !ctl.DB.MemberCreate(member) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	right := model_database.RoleDefault
	rightId := ctl.DB.RoleCreate(right)
	if rightId == nil {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}
	right.Id = *rightId

	role := model_database.RoleAssign{
		GroupId: group.Id,
		UserId:  user.Id,
		RightId: right.Id,
	}

	if !ctl.DB.RoleAssign(role) {
		return ctl.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	ctl.HTMXRedirect(controller.PathRedirectGroup(group.Id))
	return ctl.RenderSuccess(i18n.MessageSuccCreatedGroup, id)
}
