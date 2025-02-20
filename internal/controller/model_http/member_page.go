package model_http

import (
	"github.com/Mopsgamer/draqun/internal/controller/controller_http"
	"github.com/Mopsgamer/draqun/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

type MembersPage struct {
	MemberOfUriGroup
	Page uint64 `uri:"members_page"`
}

const MembersPagination uint64 = 5

func (request *MembersPage) HandleHtmx(ctl controller_http.ControllerHttp) error {
	id := "error-loading-member-list"
	if err := ctl.BindAll(request); err != nil {
		return ctl.RenderInternalError(id)
	}

	member, group, user := request.Member(ctl)
	if user == nil {
		reqLogout := UserLogout{CookieUserToken: request.CookieUserToken}
		return reqLogout.HandleHtmx(ctl)
	}

	if group == nil {
		return ctl.Ctx.SendString(i18n.MessageErrGroupNotFound)
	}

	if member == nil {
		return ctl.Ctx.SendString(i18n.MessageErrNotGroupMember)
	}

	memberList := ctl.DB.MemberListPage(request.GroupId, request.Page, MembersPagination)
	str, err := ctl.RenderString("partials/chat-members", fiber.Map{
		"GroupId":           request.GroupId,
		"MemberList":        memberList,
		"MembersPageNext":   request.Page + 1,
		"MembersPagination": MembersPagination,
	})
	if err != nil {
		return err
	}

	return ctl.Ctx.SendString(str)
}
