package logic_http

import (
	"restapp/internal/i18n"
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

	if r.DB.MemberById(*req.GroupId, user.Id) == nil {
		return r.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	memberList := r.DB.MemberListPage(*req.GroupId, req.Page, MembersPagination)
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
