package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_database"
	"restapp/internal/logic/model_request"
)

func (r LogicHTTP) GroupCreate() error {
	id := "new-group-error"
	req := new(model_request.GroupCreate)
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupByGroupname(req.Name) != nil {
		return r.RenderDanger(i18n.MessageErrGroupExistsGroupname, id)
	}

	if !model_database.IsValidGroupName(req.Name) {
		return r.RenderDanger(i18n.MessageErrGroupName, id)
	}

	if !model_database.IsValidGroupNick(req.Nick) {
		return r.RenderDanger(i18n.MessageErrGroupNick, id)
	}

	if !model_database.IsValidGroupPassword(req.Password) {
		return r.RenderDanger(i18n.MessageErrGroupPassword, id)
	}

	if !model_database.IsValidGroupDescription(req.Description) {
		return r.RenderDanger(i18n.MessageErrGroupDescription, id)
	}

	if !model_database.IsValidGroupMode(req.Mode) {
		return r.RenderDanger(i18n.MessageErrGroupMode+" Got: '"+req.Mode+"'.", id)
	}

	// TODO: validate avatar

	group := req.Group(user.Id)
	groupId := r.DB.GroupCreate(*group)
	if groupId == nil {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	group.Id = *groupId

	member := model_database.Member{
		GroupId:  group.Id,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  true,
		IsBanned: false,
	}

	if !r.DB.UserJoinGroup(member) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	right := model_database.RoleDefault
	rightId := r.DB.RoleCreate(right)
	if rightId == nil {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}
	right.Id = *rightId

	role := model_database.RoleAssign{
		GroupId: group.Id,
		UserId:  user.Id,
		RightId: right.Id,
	}

	if !r.DB.RoleAssign(role) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRedirect(logic.PathRedirectGroup(group.Id))
	return r.RenderSuccess(i18n.MessageSuccCreatedGroup, id)
}

// TODO: implement group change

func (r LogicHTTP) GroupDelete() error {
	id := "group-delete-error"
	req := new(model_request.GroupDelete)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
	}

	member := r.DB.MemberById(req.GroupId, user.Id)
	if !member.IsOwner {
		return r.RenderDanger(i18n.MessageErrNoRights, id)
	}

	if !r.DB.GroupDelete(req.GroupId) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return nil
}
