package logic_http

import (
	i18n "restapp/internal/i18n"
	"restapp/internal/logic/logic_websocket"
	"restapp/internal/logic/model"
	"restapp/internal/logic/model_request"
)

func (r LogicHTTP) GroupCreate() error {
	id := "new-group-error"
	req := new(model_request.GroupCreate)
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupByGroupname(req.Name) != nil {
		return r.RenderDanger(i18n.MessageErrGroupExistsGroupname, id)
	}

	if !model.IsValidGroupName(req.Name) {
		return r.RenderDanger(i18n.MessageErrGroupName, id)
	}

	if !model.IsValidGroupNick(req.Nick) {
		return r.RenderDanger(i18n.MessageErrGroupNick, id)
	}

	if !model.IsValidGroupPassword(req.Password) {
		return r.RenderDanger(i18n.MessageErrGroupPassword, id)
	}

	if !model.IsValidGroupDescription(req.Description) {
		return r.RenderDanger(i18n.MessageErrGroupDescription, id)
	}

	if !model.IsValidGroupMode(req.Mode) {
		return r.RenderDanger(i18n.MessageErrGroupMode+" Got: '"+req.Mode+"'.", id)
	}

	// TODO: validate avatar

	group := req.Group(user.Id)
	groupId := r.DB.GroupCreate(*group)
	if groupId == nil {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	member := &model.Member{
		GroupId:  *groupId,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  true,
		IsBanned: false,
	}
	if r.GroupJoin(*member) == nil {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRedirect(group.PagePath())
	return r.RenderSuccess(i18n.MessageSuccCreatedGroup, id)
}

// TODO: implement group change

func (r LogicHTTP) GroupDelete() error {
	id := "group-delete-error"
	req := new(model_request.GroupDelete)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	member := r.DB.GroupMemberById(req.GroupId, user.Id)
	if !member.IsOwner {
		return r.RenderDanger(i18n.MessageErrNoRights, id)
	}

	if !r.DB.GroupDelete(req.GroupId) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return nil
}

func (r LogicHTTP) GroupJoin(member model.Member) *uint64 {
	for _, ws := range (*logic_websocket.WebsocketConnections.Users)[member.UserId] {
		ws.WebsocketRender("partials/group-member", member)
	}

	return r.DB.GroupMemberCreate(member)
}

func (r LogicHTTP) GroupLeave() error {
	id := "group-leave-error"
	req := new(model_request.GroupLeave)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupMemberById(req.GroupId, user.Id) == nil {
		return r.RenderDanger(i18n.MessageErrNotGroupMember, id)
	}

	r.HTMXRedirect("/chat")
	return r.RenderSuccess(i18n.MessageSuccLeavedGroup, id)
}

func (r LogicHTTP) MessageSend(message model.Message) *uint64 {
	for _, ws := range (*logic_websocket.WebsocketConnections.Users)[message.AuthorId] {
		ws.WebsocketRender("partials/message", message)
	}

	return r.DB.MessageCreate(message)
}
