package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"
)

func (r Responder) GroupCreate() error {
	id := "group-create-error"
	req := new(model_request.GroupCreate)
	if err := r.Bind().Form(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
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

	// TODO: validate avatar and mode

	group := req.Group(user.Id)
	r.DB.GroupCreate(*group)

	r.HTMXRedirect(group.PagePath())
	return r.RenderSuccess(MessageSuccCreatedGroup, id)
}

func (r Responder) GroupDelete() error {
	id := "group-delete-error"
	req := new(model_request.GroupDelete)
	if err := r.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
	}

	member := r.DB.GroupMember(req.GroupId, user.Id)
	if !member.IsOwner {
		return r.RenderDanger(MessageErrNoRights, id)
	}

	if !r.DB.GroupDelete(req.GroupId) {
		return r.RenderDanger(MessageFatalDatabaseQuery, id)
	}

	r.HTMXRefresh()
	return nil
}

func (r Responder) GroupLeave() error {
	id := "group-leave-error"
	req := new(model_request.GroupLeave)
	if err := r.Bind().URI(req); err != nil {
		return r.RenderDanger(MessageErrInvalidRequest, id)
	}

	user := r.User()
	if user == nil {
		return nil
	}

	if r.DB.GroupMember(req.GroupId, user.Id) == nil {
		return r.RenderDanger(MessageErrNotGroupMember, id)
	}

	r.HTMXRedirect("/chat")
	return r.RenderSuccess(MessageSuccLeavedGroup, id)
}
