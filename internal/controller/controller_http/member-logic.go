package controller_http

import (
	"restapp/internal/controller"
	"restapp/internal/controller/model_database"
	"restapp/internal/controller/model_http"
	"restapp/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

const MembersPagination uint64 = 5

func (r ControllerHttp) MembersPage() error {
	req := new(model_http.MembersPage)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderInternalError("error-loading-member-list")
	}

	member, _, group := r.Member()

	if group == nil {
		return r.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
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

func (r ControllerHttp) GroupLeave() error {
	id := "group-leave-error"
	req := new(model_http.GroupLeave)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderInternalError(id)
	}

	member, _, group := r.Member()

	if group == nil {
		return r.RenderDanger(i18n.MessageErrGroupNotFound, id)
	}

	if member == nil {
		return r.RenderDanger(i18n.MessageErrNotGroupMember, id)
	}

	// TODO: do not leave if last owner and there are other non-owner members.
	// Ask for assigning new owner before leave.

	// TODO: delete group on leave if no other members.

	r.HTMXRedirect("/chat")
	return r.RenderSuccess(i18n.MessageSuccLeavedGroup, id)
}

func (r ControllerHttp) GroupJoin() error {
	id := "group-join-error"
	req := new(model_http.GroupJoin)
	if err := r.Ctx.Bind().URI(req); err != nil {
		return r.RenderInternalError(id)
	}

	member, user, group := r.Member()
	if group == nil {
		return r.RenderDanger(i18n.MessageErrGroupNotFound, id)
	}

	if member != nil {
		return r.RenderDanger(i18n.MessageErrAlreadyGroupMember, id)
	}

	member = &model_database.Member{
		GroupId:  group.Id,
		UserId:   user.Id,
		Nick:     nil,
		IsOwner:  false,
		IsBanned: false,
	}

	if !r.DB.UserJoinGroup(*member) {
		return r.RenderDanger(i18n.MessageFatalDatabaseQuery, id)
	}

	if len(r.DB.MemberRoleList(group.Id, user.Id)) < 1 {
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
	}

	r.HTMXRedirect(controller.PathRedirectGroup(req.GroupId))
	return r.RenderSuccess(i18n.MessageSuccJoinedGroup, id)
}
