package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"
)

func (r Responder) GroupCreate() error {
	id := "new-group-error"
	req := new(model_request.GroupCreate)
	if err := r.Ctx.Bind().Form(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupByGroupname(req.Name) != nil {
		return r.RenderDanger(MessageErrGroupExistsGroupname, id)
	}

	if !model.IsValidGroupName(req.Name) {
		return r.RenderDanger(MessageErrGroupName, id)
	}

	if !model.IsValidGroupNick(req.Nick) {
		return r.RenderDanger(MessageErrGroupNick, id)
	}

	if !model.IsValidGroupPassword(req.Password) {
		return r.RenderDanger(MessageErrGroupPassword, id)
	}

	if !model.IsValidGroupDescription(req.Description) {
		return r.RenderDanger(MessageErrGroupDescription, id)
	}

	if !model.IsValidGroupMode(req.Mode) {
		return r.RenderDanger(MessageErrGroupMode+" Got: '"+req.Mode+"'.", id)
	}

	// TODO: validate avatar

	group := req.Group(user.Id)
	groupId := r.DB.GroupCreate(*group)
	if groupId == nil {
		return r.RenderDanger(MessageFatalDatabaseQuery, id)
	}

	member := &model.Member{
		GroupId:  *groupId,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  true,
		IsBanned: false,
	}
	if r.GroupJoin(*member) == nil {
		return r.RenderDanger(MessageFatalDatabaseQuery, id)
	}

	r.HTMXRedirect(group.PagePath())
	return r.RenderSuccess(MessageSuccCreatedGroup, id)
}

// TODO: implement group change

func (r Responder) GroupDelete() error {
	id := "group-delete-error"
	req := new(model_request.GroupDelete)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	member := r.DB.GroupMemberById(req.GroupId, user.Id)
	if !member.IsOwner {
		return r.RenderDanger(MessageErrNoRights, id)
	}

	if !r.DB.GroupDelete(req.GroupId) {
		return r.RenderDanger(MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return nil
}

func (r Responder) GroupJoin(member model.Member) *uint64 {
	for _, ws := range (*WebsocketConnections.Users)[member.UserId] {
		ws.WebsocketRender("partials/group-member", member)
	}

	return r.DB.GroupMemberCreate(member)
}

func (r Responder) GroupLeave() error {
	id := "group-leave-error"
	req := new(model_request.GroupLeave)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupMemberById(req.GroupId, user.Id) == nil {
		return r.RenderDanger(MessageErrNotGroupMember, id)
	}

	r.HTMXRedirect("/chat")
	return r.RenderSuccess(MessageSuccLeavedGroup, id)
}

func (r Responder) MessageSend(message model.Message) *uint64 {
	for _, ws := range (*WebsocketConnections.Users)[message.AuthorId] {
		ws.WebsocketRender("partials/message", message)
	}

	return r.DB.MessageCreate(message)
}
