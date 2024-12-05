package logic_http

import (
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/model_request"

	"github.com/gofiber/fiber/v3"
)

const MembersPagination uint64 = 5

func (r LogicHTTP) MembersPage() error {
	req := new(model_request.MembersPage)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.Ctx.SendString(i18n.MessageErrInvalidRequest)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.MemberById(req.GroupId, user.Id) == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	memberList := r.DB.MemberListPage(req.GroupId, req.Page, MembersPagination)
	str, err := r.RenderString("partials/chat-members", fiber.Map{
		"GroupId":           req.GroupId,
		"MemberList":        memberList,
		"MembersPageNext":   req.Page + 1,
		"MembersPagination": MembersPagination,
	})
	if err != nil {
		return err
	}

	return r.Ctx.SendString(str)
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

	if r.DB.MemberById(req.GroupId, user.Id) == nil {
		return r.RenderDanger(i18n.MessageErrNotGroupMember, id)
	}

	// FIXME: do not leave if last owner and there are other non-owner members.
	// Ask for assigning new owner before leave.

	// FIXME: delete group on leave if no other members.

	r.HTMXRedirect("/chat")
	return r.RenderSuccess(i18n.MessageSuccLeavedGroup, id)
}

func (r LogicHTTP) GroupJoin() error {
	id := "group-join-error"
	req := new(model_request.GroupJoin)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
	}

	user, _ := r.User()
	if user == nil {
		return nil
	}

	if r.DB.MemberById(req.GroupId, user.Id) != nil {
		return r.RenderDanger(i18n.MessageErrAlreadyGroupMember, id)
	}

	r.HTMXRedirect(logic.PathRedirectGroup(req.GroupId))
	return r.RenderSuccess(i18n.MessageSuccJoinedGroup, id)
}
