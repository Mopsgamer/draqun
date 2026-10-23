package routes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/Mopsgamer/draqun/server/render"
	"github.com/Mopsgamer/draqun/server/session"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

type GroupCreate struct {
	Name        model.Name             `form:"name"`
	Nick        model.Moniker          `form:"nick"`
	Password    model.OptionalPassword `form:"password"`
	Mode        model.GroupMode        `form:"mode"`
	Description model.Description      `form:"description"`
	Avatar      model.Avatar           `form:"avatar"`
}

type PermissionCreateForm struct {
	Name  model.PermName  `form:"name"`
	Level model.PermLevel `form:"level"`
}

type PermissionUpdateForm struct {
	Name  model.PermName  `form:"name"`
	Level model.PermLevel `form:"level"`
}

type PermissionAttachForm struct {
	EntityType string `form:"entity_type" json:"entity_type"`
	EntityId   uint64 `form:"entity_id" json:"entity_id"`
}

type EntitySearchForm struct {
	PermId     uint64 `form:"perm_id" query:"perm_id"`
	EntityType string `form:"entity_type" query:"entity_type"`
	Q          string `form:"q" query:"q"`
}

type RoleCreate struct {
	Name    model.Name    `form:"name"`
	Moniker model.Moniker `form:"moniker"`
	Color   string        `form:"color"`
}

type RoleChange struct {
	Name    model.Name    `form:"name"`
	Moniker model.Moniker `form:"moniker"`
	Color   string        `form:"color"`
}

type RoleAssignForm struct {
	RoleId uint64 `form:"role_id"`
}

type MemberKickForm struct {
	Description model.Description `form:"description"`
}

type MemberBanForm struct {
	Description  model.Description `form:"description"`
	DurationDays uint              `form:"duration_days"`
}

type MessageCreate struct {
	Content   string `form:"content"`
	ChannelId uint64 `form:"channel_id"`
}

type TicketCreateForm struct {
	Title     string   `form:"title"`
	Content   string   `form:"content"`
	ParentIds []uint64 `form:"parent_ids"`
}

type TicketStatusForm struct {
	Status        string `form:"status"`
	ClosedReason  string `form:"closed_reason"`
	DuplicateOfId uint64 `form:"duplicate_of_id"`
}

type TicketEditForm struct {
	Title            string   `form:"title"`
	Content          string   `form:"content"`
	ParentIds        []uint64 `form:"parent_ids"`
	RemoveParentFlag bool     `form:"remove_parent"`
}

type TicketAssignForm struct {
	UserId uint64 `form:"user_id"`
}

type MemberSearchForm struct {
	Q         string `form:"q" query:"q"`
	RoleId    uint64 `form:"role_id" query:"role_id"`
	ChannelId uint64 `form:"channel_id" query:"channel_id"`
}

type TicketSearchForm struct {
	Q         string `form:"q" query:"q"`
	ExcludeId uint64 `form:"exclude_id" query:"exclude_id"`
	Mode      string `form:"mode" query:"mode"`
}

type ChannelCreate struct {
	Name        model.Name        `form:"name"`
	Moniker     model.Moniker     `form:"moniker"`
	Description model.Description `form:"description"`
	Type        model.ChannelType `form:"type"`
	ParentId    uint64            `form:"parent_id"`
}

type ChannelChange struct {
	Name        model.Name        `form:"name"`
	Moniker     model.Moniker     `form:"moniker"`
	Description model.Description `form:"description"`
	Type        model.ChannelType `form:"type"`
}

type GroupChange struct {
	Name        model.Name             `form:"name"`
	Nick        model.Moniker          `form:"nick"`
	Password    model.OptionalPassword `form:"password"`
	Mode        model.GroupMode        `form:"mode"`
	Description model.Description      `form:"description"`
	Avatar      model.Avatar           `form:"avatar"`
}

func RouteGroups(app *fiber.App) fiber.Router {
	return app.Group("/groups").
		Put("/:group_id/join",
			perms.UserByAuth(),
			perms.GroupById("group_id"),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				if group.IsBanned(user.Id) {
					return htmx.AlertGroupMemberNotAllowed
				}

				perms.MemberByAuthAndGroupIdFromCtx(ctx, "group_id")
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)
				if member.IsEmpty() {
					// never been a member
					member = model.NewMember(group.Id, user.Id, "")
					if err := member.Insert(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				} else if member.IsDeleted {
					// been a member, now isn't
					member.IsDeleted = false
					if err := member.Update(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				} else {
					// already a member
					return htmx.AlertUseless
				}

				if err := member.JoinActed(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				broadcastGroupMembersUpdate(app, group)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, group.Url(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/permissions",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[PermissionCreateForm](),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				request := fiber.Locals[PermissionCreateForm](ctx, perms.LocalForm)

				existingPerms, _ := model.GetGroupPermissions(group.Id)
				maxPriority := 0
				for _, ep := range existingPerms {
					if ep.Priority > maxPriority {
						maxPriority = ep.Priority
					}
				}

				p := model.Permission{
					GroupId:  group.Id,
					Name:     request.Name,
					Level:    request.Level,
					Priority: maxPriority + 10,
				}

				if err := p.Validate(); err != nil {
					return err
				}

				if err := p.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/:group_id/permissions/:perm_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[PermissionUpdateForm](),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				permId := fiber.Params[uint64](ctx, "perm_id")
				request := fiber.Locals[PermissionUpdateForm](ctx, perms.LocalForm)

				perm, err := model.NewPermissionFromId(permId, group.Id)
				if err != nil || perm.IsEmpty() {
					return htmx.AlertGroupNotFound
				}

				perm.Name = request.Name
				perm.Level = request.Level

				if err := perm.Validate(); err != nil {
					return err
				}

				if err := perm.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/permissions/:perm_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				permId := fiber.Params[uint64](ctx, "perm_id")

				perm, err := model.NewPermissionFromId(permId, group.Id)
				if err != nil || perm.IsEmpty() {
					return htmx.AlertGroupNotFound
				}

				if err := perm.Delete(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/permissions/:perm_id/move",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				permId := fiber.Params[uint64](ctx, "perm_id")
				direction := ctx.Query("direction", ctx.FormValue("direction"))

				permsList, err := model.GetGroupPermissions(group.Id)
				if err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				targetIdx := -1
				for i, p := range permsList {
					if p.Id == permId {
						targetIdx = i
						break
					}
				}

				if targetIdx == -1 {
					return htmx.AlertGroupNotFound
				}

				swapIdx := -1
				if direction == "up" && targetIdx > 0 {
					swapIdx = targetIdx - 1
				} else if direction == "down" && targetIdx < len(permsList)-1 {
					swapIdx = targetIdx + 1
				}

				if swapIdx != -1 {
					for i := range permsList {
						permsList[i].Priority = (len(permsList) - i) * 10
						_ = permsList[i].Update()
					}

					permsList[targetIdx].Priority, permsList[swapIdx].Priority = permsList[swapIdx].Priority, permsList[targetIdx].Priority
					_ = permsList[targetIdx].Update()
					_ = permsList[swapIdx].Update()
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/permissions/:perm_id/attach",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[PermissionAttachForm](),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				permId := fiber.Params[uint64](ctx, "perm_id")
				request := fiber.Locals[PermissionAttachForm](ctx, perms.LocalForm)

				perm, err := model.NewPermissionFromId(permId, group.Id)
				if err != nil || perm.IsEmpty() {
					return htmx.AlertGroupNotFound
				}

				var attachErr error
				switch request.EntityType {
				case "channel":
					attachErr = perm.AddChannel(request.EntityId)
				case "role":
					attachErr = perm.AddRole(request.EntityId)
				case "user":
					attachErr = perm.AddUser(request.EntityId)
				}
				if attachErr != nil {
					return htmx.AlertDatabase.Join(attachErr)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/permissions/:perm_id/detach/:entity_type/:entity_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				permId := fiber.Params[uint64](ctx, "perm_id")
				entityType := fiber.Params[string](ctx, "entity_type")
				entityId := fiber.Params[uint64](ctx, "entity_id")

				perm, err := model.NewPermissionFromId(permId, group.Id)
				if err != nil || perm.IsEmpty() {
					return htmx.AlertGroupNotFound
				}

				switch entityType {
				case "channel":
					_ = perm.RemoveChannel(entityId)
				case "role":
					_ = perm.RemoveRole(entityId)
				case "user":
					_ = perm.RemoveUser(entityId)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/permissions", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/:group_id/permissions/entities/search",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[EntitySearchForm](),
			func(ctx fiber.Ctx) error {
				return handleEntitySearch(ctx)
			},
		).
		Post("/:group_id/permissions/entities/search",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[EntitySearchForm](),
			func(ctx fiber.Ctx) error {
				return handleEntitySearch(ctx)
			},
		).
		Post("/:group_id/roles/:role_id/move",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				roleId := fiber.Params[uint64](ctx, "role_id")
				direction := ctx.FormValue("direction")
				if direction == "" {
					direction = ctx.Query("direction", "up")
				}

				roles := group.Roles()
				targetIdx := -1
				for i, r := range roles {
					if r.Id == roleId {
						targetIdx = i
						break
					}
				}

				if targetIdx == -1 {
					return htmx.AlertGroupMemberNotFound
				}

				swapIdx := -1
				if direction == "up" && targetIdx > 0 {
					swapIdx = targetIdx - 1
				} else if direction == "down" && targetIdx < len(roles)-1 {
					swapIdx = targetIdx + 1
				}

				if swapIdx != -1 {
					for i := range roles {
						roles[i].Position = uint(i)
						_ = roles[i].Update()
					}

					roles[targetIdx].Position, roles[swapIdx].Position = roles[swapIdx].Position, roles[targetIdx].Position
					_ = roles[targetIdx].Update()
					_ = roles[swapIdx].Update()

					actionRole := model.ActionRole{
						RoleId:      roleId,
						CreatorId:   caller.Id,
						GroupId:     group.Id,
						Action:      "update",
						Description: model.Description(fmt.Sprintf("Reordered role %s", roles[targetIdx].Moniker)),
						ActedAt:     model.TimePast(time.Now()),
					}
					_ = actionRole.Insert()
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/:group_id/change",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			perms.UseBind[GroupChange](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[GroupChange](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				file, errFile := ctx.FormFile("avatar")
				if errFile == nil && file != nil {
					if err := os.MkdirAll("uploads", 0755); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
					ext := filepath.Ext(file.Filename)
					if ext == "" {
						ext = ".png"
					}
					filename := fmt.Sprintf("group-%d-%d%s", group.Id, time.Now().UnixNano(), ext)
					dst := filepath.Join("uploads", filename)
					if err := ctx.SaveFile(file, dst); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
					group.Avatar = model.Avatar("/uploads/" + filename)
				}

				hasChanges := group.Moniker != request.Nick ||
					group.Name != request.Name ||
					group.Description != request.Description ||
					group.Mode != request.Mode ||
					group.Password.Compare(request.Password) != nil ||
					errFile == nil

				if !hasChanges {
					return htmx.AlertUseless
				}

				group.Moniker = request.Nick
				group.Name = request.Name
				group.Description = request.Description
				group.Mode = model.GroupMode(request.Mode)
				var err error
				group.Password, err = request.Password.Hash()
				if err != nil {
					return htmx.AlertEncryption.Join(err)
				}

				if err := group.Validate(); err != nil {
					return err
				}

				if err := group.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, htmx.Path(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/star",
			perms.RequirePerm("group_id", model.PermNameRead),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				isStarred := model.IsChannelStarred(user.Id, channel.Id)
				if isStarred {
					_ = model.RemoveStarredChannel(user.Id, channel.Id)
					isStarred = false
				} else {
					_ = model.AddStarredChannel(user.Id, channel.Id)
					isStarred = true
				}

				if htmx.IsHtmx(ctx) {
					var iconClass string
					var titleText string
					if isStarred {
						iconClass = "icon icon-[lucide--star] size-4 text-amber-400 fill-amber-400"
						titleText = "Unstar channel"
					} else {
						iconClass = "icon icon-[lucide--star] size-4 text-base-content/70"
						titleText = "Star channel"
					}
					html := fmt.Sprintf(`<button hx-post="/groups/%d/channels/%d/star" hx-swap="outerHTML" class="btn btn-ghost btn-square btn-sm text-base-content/70 hover:text-base-content" aria-label="%s" title="%s"><span class="%s" aria-hidden="true"></span></button>`,
						group.Id, channel.Id, titleText, titleText, iconClass)
					return ctx.Type("html").SendString(html)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/move",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")
				direction := ctx.Query("direction", "up")

				channels, err := model.GetGroupChannels(group.Id)
				if err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				targetIdx := -1
				for i, ch := range channels {
					if ch.Id == channelId {
						targetIdx = i
						break
					}
				}

				if targetIdx == -1 {
					return htmx.AlertGroupNotFound
				}

				swapIdx := -1
				if direction == "up" && targetIdx > 0 {
					swapIdx = targetIdx - 1
				} else if direction == "down" && targetIdx < len(channels)-1 {
					swapIdx = targetIdx + 1
				}

				if swapIdx != -1 {
					for i, ch := range channels {
						_ = model.UpdateGroupChannelPosition(group.Id, ch.Id, uint(i))
					}

					_ = model.UpdateGroupChannelPosition(group.Id, channels[targetIdx].Id, uint(swapIdx))
					_ = model.UpdateGroupChannelPosition(group.Id, channels[swapIdx].Id, uint(targetIdx))
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/channels", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/create",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			perms.UseBind[ChannelCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[ChannelCreate](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				chType := request.Type
				if chType == "" {
					chType = model.ChannelTypeText
				}
				channel := model.NewChannel(request.Name, request.Moniker, request.Description, chType)
				if err := channel.Validate(); err != nil {
					return err
				}

				if err := channel.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				groupChannels, _ := model.GetGroupChannels(group.Id)
				pos := uint(len(groupChannels))
				_ = model.AddGroupChannel(group.Id, channel.Id, pos)

				if request.ParentId != 0 {
					_ = model.AddChannelRelation(request.ParentId, channel.Id, "subchannel")
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/channels", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/:group_id/channels/:channel_id",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			perms.UseBind[ChannelChange](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[ChannelChange](ctx, perms.LocalForm)
				channelId := fiber.Params[uint64](ctx, "channel_id")
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				if request.Name != "" {
					channel.Name = request.Name
				}
				if request.Moniker != "" {
					channel.Moniker = request.Moniker
				}
				if request.Description != "" {
					channel.Description = request.Description
				}
				if request.Type != "" {
					channel.Type = request.Type
				}

				if err := channel.Validate(); err != nil {
					return err
				}

				if err := channel.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/channels", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/channels/:channel_id",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				channelId := fiber.Params[uint64](ctx, "channel_id")
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				_ = channel.Delete()
				_ = model.RemoveGroupChannel(group.Id, channelId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/channels", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/:group_id/members/search",
			perms.RequirePerm("group_id", model.PermNameMembers),
			perms.UseBind[MemberSearchForm](),
			func(ctx fiber.Ctx) error {
				return handleMemberSearch(ctx)
			},
		).
		Post("/:group_id/members/search",
			perms.RequirePerm("group_id", model.PermNameMembers),
			perms.UseBind[MemberSearchForm](),
			func(ctx fiber.Ctx) error {
				return handleMemberSearch(ctx)
			},
		).
		Post("/:group_id/channels/:channel_id/assignees",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[TicketAssignForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[TicketAssignForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				if request.UserId != 0 {
					if model.IsChannelAssignee(channel.Id, request.UserId) {
						return htmx.AlertUseless
					}
					_ = model.AddChannelAssignee(channel.Id, request.UserId)
					caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
					notif := model.NewNotification(
						request.UserId,
						caller.Id,
						group.Id,
						channel.Id,
						0,
						"ticket_assigned",
						fmt.Sprintf("You were assigned to ticket #%s in %s", channel.Moniker, group.Moniker),
						fmt.Sprintf("%s assigned you to ticket #%s", caller.Moniker, channel.Moniker),
					)
					_ = model.CreateAndPushNotification(app, notif)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, channel.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/channels/:channel_id/assignees/:user_id",
			perms.RequirePerm("group_id", model.PermNameWrite),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")
				targetUserId := fiber.Params[uint64](ctx, "user_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				_ = model.RemoveChannelAssignee(channel.Id, targetUserId)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)

				notif := model.NewNotification(
					targetUserId,
					caller.Id,
					group.Id,
					channel.Id,
					0,
					"ticket_unassigned",
					fmt.Sprintf("You were unassigned from ticket #%s in %s", channel.Moniker, group.Moniker),
					fmt.Sprintf("%s unassigned you from ticket #%s", caller.Moniker, channel.Moniker),
				)
				_ = model.CreateAndPushNotification(app, notif)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, channel.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/:group_id/tickets/search",
			perms.RequirePerm("group_id", model.PermNameRead),
			perms.UseBind[TicketSearchForm](),
			func(ctx fiber.Ctx) error {
				return handleTicketSearch(ctx)
			},
		).
		Post("/:group_id/tickets/search",
			perms.RequirePerm("group_id", model.PermNameRead),
			perms.UseBind[TicketSearchForm](),
			func(ctx fiber.Ctx) error {
				return handleTicketSearch(ctx)
			},
		).
		Post("/:group_id/channels/:parent_id/tickets/create",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[TicketCreateForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[TicketCreateForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				parentId := fiber.Params[uint64](ctx, "parent_id")

				parentChannel, err := model.NewChannelFromId(parentId)
				if err != nil || !parentChannel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				if len(strings.TrimSpace(request.Title)) == 0 {
					return htmx.AlertFormatMoniker
				}

				channel := model.NewChannel(model.Name("ticket_0"), model.Moniker(request.Title), model.Description(request.Content), model.ChannelTypeTicket)
				channel.AuthorId = user.Id
				channel.Status = "open"

				if err := channel.Validate(); err != nil {
					return err
				}

				if err := channel.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				channel.Name = model.Name(fmt.Sprintf("ticket_%d", channel.Id))
				if err := channel.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				parentIds := request.ParentIds
				hasMainParent := false
				for _, pid := range parentIds {
					if pid == parentId {
						hasMainParent = true
						break
					}
				}
				if !hasMainParent {
					parentIds = append(parentIds, parentId)
				}
				_ = model.SetParentChannels(channel.Id, parentIds)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, channel.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/tickets/edit",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[TicketEditForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[TicketEditForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				if len(strings.TrimSpace(request.Title)) > 0 {
					channel.Moniker = model.Moniker(request.Title)
				}
				channel.Description = model.Description(request.Content)
				channel.UpdatedAt = model.TimePast(time.Now())

				if err := channel.Validate(); err != nil {
					return err
				}

				if len(request.ParentIds) == 0 {
					return htmx.AlertFormatMoniker
				}
				if err := model.SetParentChannels(channel.Id, request.ParentIds); err != nil {
					return err
				}

				if err := channel.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, channel.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/status",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[TicketStatusForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[TicketStatusForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				if request.Status == "open" {
					channel.Status = "open"
					channel.ClosedReason = ""
					channel.DuplicateOfId = 0
				} else if request.Status == "closed" {
					channel.Status = "closed"
					reason := request.ClosedReason
					if reason != "not_planned" && reason != "duplicate" {
						reason = "completed"
					}
					channel.ClosedReason = reason
					if reason == "duplicate" {
						channel.DuplicateOfId = request.DuplicateOfId
					} else {
						channel.DuplicateOfId = 0
					}
				}

				if err := channel.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/chat/groups/%d/channels/%d", group.Id, channel.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/dm/channels/:channel_id/messages/create",
			perms.UserByAuth(),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				if !model.IsUserInDirectChannel(channelId, user.Id) {
					return htmx.AlertGroupMemberNotAllowed
				}

				message := model.NewMessageFilled(channelId, 0, user.Id, request.Content)

				if err := saveMessageAttachment(ctx, 0, &message); err != nil {
					return err
				}

				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				// Push notification & message update to the recipient in DM
				otherUser, errOther := model.GetDirectChannelOtherUser(channelId, user.Id)
				if errOther == nil && otherUser != nil && !otherUser.IsEmpty() {
					notif := model.NewNotification(
						otherUser.Id,
						user.Id,
						0,
						channelId,
						message.Id,
						"dm",
						fmt.Sprintf("Direct Message from %s", user.Moniker),
						string(message.Content),
					)
					_ = model.CreateAndPushNotification(app, notif)

					bufMsg, errBuf := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if errBuf == nil {
						msgStr := bufMsg.String()
						oobMsg := render.WrapOob(fmt.Sprintf("beforeend:#chat-messages-%d", channelId), &msgStr)
						session.UserSessionMap.PushUser(otherUser.Id, oobMsg)
					}

					PushHomeChannelsOob(app, otherUser.Id)
				}

				model.ProcessMessageMentions(app, message)

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()
					oobMsg := render.WrapOob(fmt.Sprintf("beforeend:#chat-messages-%d", channelId), &str)

					return ctx.Type("html").SendString(oobMsg)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/messages/create",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				message := model.NewMessageFilled(channelId, group.Id, user.Id, request.Content)

				if err := saveMessageAttachment(ctx, group.Id, &message); err != nil {
					return err
				}

				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				model.ProcessMessageMentions(app, message)

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()
					oobMsg := render.WrapOob(fmt.Sprintf("beforeend:#chat-messages-%d", channelId), &str)

					session.UserSessionMap.PushGroup(
						group.Id,
						oobMsg,
					)

					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/messages/create",
			perms.RequirePerm("group_id", model.PermNameWrite),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				channelId := request.ChannelId
				if channelId == 0 {
					channels, _ := model.GetAllGroupChannels(group.Id)
					if len(channels) > 0 {
						channelId = channels[0].Id
					} else {
						return htmx.AlertGroupNotFound
					}
				}

				message := model.NewMessageFilled(channelId, group.Id, user.Id, request.Content)

				if err := saveMessageAttachment(ctx, group.Id, &message); err != nil {
					return err
				}

				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				model.ProcessMessageMentions(app, message)

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()
					oobMsg := render.WrapOob(fmt.Sprintf("beforeend:#chat-messages-%d", channelId), &str)

					session.UserSessionMap.PushGroup(
						group.Id,
						oobMsg,
					)

					return ctx.SendStatus(fiber.StatusOK)
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/create",
			perms.UserByAuth(),
			perms.UseBind[GroupCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[GroupCreate](ctx, perms.LocalForm)
				existingGroup, _ := model.NewGroupFromName(request.Name)
				if !existingGroup.IsEmpty() {
					return htmx.AlertGroupExistsName
				}

				hash, err := request.Password.Hash()
				if err != nil {
					return htmx.AlertEncryption.Join(err)
				}

				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				avatarPath := request.Avatar
				file, errFile := ctx.FormFile("avatar")
				if errFile == nil && file != nil {
					if err := os.MkdirAll("uploads", 0755); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
					ext := filepath.Ext(file.Filename)
					if ext == "" {
						ext = ".png"
					}
					filename := fmt.Sprintf("group-new-%d%s", time.Now().UnixNano(), ext)
					dst := filepath.Join("uploads", filename)
					if err := ctx.SaveFile(file, dst); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
					avatarPath = model.Avatar("/uploads/" + filename)
				}

				group := model.NewGroup(user.Id, request.Nick, request.Name, request.Mode, hash, request.Description, avatarPath)
				if err := group.Validate(); err != nil {
					return err
				}

				if err := group.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				ctx.Locals(perms.LocalGroup, group)

				member := model.NewMember(group.Id, user.Id, "")
				if err := member.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				everyone := model.NewRoleEveryone(group.Id)
				if err := everyone.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, group.Url(ctx))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/dm/channels/:channel_id/messages/page/:messages_page",
			perms.UserByAuth(),
			func(ctx fiber.Ctx) error {
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				if !model.IsUserInDirectChannel(channelId, user.Id) {
					return htmx.AlertGroupMemberNotAllowed
				}

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				page := fiber.Params[uint](ctx, "messages_page")
				const MessagesPagination uint = 5
				messageList := channel.MessagesPage(page, MessagesPagination)

				if htmx.IsHtmx(ctx) {
					bind := fiber.Map{
						"MessageList":        messageList,
						"MessagesPage":       page,
						"MessagesPagination": MessagesPagination,
						"ActiveChannel":      channel,
					}

					return ctx.Render("partials/chat-messages", bind)
				}

				return ctx.JSON(messageList)
			},
		).
		Get("/:group_id/channels/:channel_id/messages/page/:messages_page",
			perms.RequirePerm("group_id", model.PermNameRead),
			func(ctx fiber.Ctx) error {
				channelId := fiber.Params[uint64](ctx, "channel_id")

				channel, err := model.NewChannelFromId(channelId)
				if err != nil || !channel.IsAvailable() {
					return htmx.AlertGroupNotFound
				}

				page := fiber.Params[uint](ctx, "messages_page")
				const MessagesPagination uint = 5
				messageList := channel.MessagesPage(page, MessagesPagination)

				if htmx.IsHtmx(ctx) {
					bind := fiber.Map{
						"MessageList":        messageList,
						"MessagesPage":       page,
						"MessagesPagination": MessagesPagination,
						"ActiveChannel":      channel,
					}

					return ctx.Render("partials/chat-messages", bind)
				}

				return ctx.JSON(messageList)
			},
		).
		Get("/:group_id/messages/page/:messages_page",
			perms.RequirePerm("group_id", model.PermNameRead),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				page := fiber.Params[uint](ctx, "messages_page")
				const MessagesPagination uint = 5

				channels, _ := model.GetAllGroupChannels(group.Id)
				var messageList []model.Message
				var activeChannel model.Channel

				if len(channels) > 0 {
					activeChannel = channels[0]
					messageList = activeChannel.MessagesPage(page, MessagesPagination)
				} else {
					return htmx.AlertGroupNotFound
				}

				if htmx.IsHtmx(ctx) {
					bind := fiber.Map{
						"MessageList":        messageList,
						"MessagesPage":       page,
						"MessagesPagination": MessagesPagination,
						"ActiveChannel":      activeChannel,
					}

					return ctx.Render("partials/chat-messages", bind)
				}

				return ctx.JSON(messageList)
			},
		).
		Get("/:group_id/members/page/:members_page",
			perms.RequirePerm("group_id", model.PermNameMembers),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				page := fiber.Params[uint](ctx, "members_page")
				const MembersPagination uint = 5
				memberList := group.UsersPage(page, MembersPagination)

				if htmx.IsHtmx(ctx) {
					bind := fiber.Map{
						"Group":             group,
						"MemberList":        memberList,
						"MembersPage":       page,
						"MembersPagination": MembersPagination,
					}

					return ctx.Render("partials/chat-members", bind)
				}

				return ctx.JSON(memberList)
			},
		).
		Delete("/:group_id/leave",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return true
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				member := fiber.Locals[model.Member](ctx, perms.LocalMember)

				isAlone := group.MembersCount() == 1
				if group.OwnerId == member.UserId && !isAlone {
					return htmx.AlertGroupMemberIsOnlyOwner
				}

				if isAlone {
					group.IsDeleted = true
					if err := group.Update(); err != nil {
						return htmx.AlertDatabase.Join(err)
					}
				}

				member.IsDeleted = true
				if err := member.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}
				if err := member.LeaveActed(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				broadcastGroupMemberLeave(app, group, member.UserId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, "/chat")
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id",
			perms.RequirePermLevel("group_id", model.PermNameSettings, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				group.IsDeleted = true
				if err := group.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.EnableRefresh(ctx)
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/roles/create",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[RoleCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[RoleCreate](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)

				existingRole, _ := model.NewRoleFromName(request.Name, group.Id)
				if !existingRole.IsEmpty() {
					return htmx.AlertGroupExistsName
				}

				role := model.Role{
					GroupId: group.Id,
					Name:    request.Name,
					Moniker: request.Moniker,
					Color:   model.ParseColorHex(request.Color),
				}

				if err := role.Validate(); err != nil {
					return err
				}

				if err := role.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				actionRole := model.ActionRole{
					RoleId:      role.Id,
					CreatorId:   caller.Id,
					GroupId:     group.Id,
					Action:      "create",
					Description: model.Description(fmt.Sprintf("Created role %s (%s)", role.Moniker, role.Name)),
					ActedAt:     model.TimePast(time.Now()),
				}
				_ = actionRole.Insert()

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Put("/:group_id/roles/:role_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[RoleChange](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[RoleChange](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				roleId := fiber.Params[uint64](ctx, "role_id")

				role, err := model.NewRoleFromId(roleId, group.Id)
				if err != nil || role.IsEmpty() {
					return htmx.AlertGroupMemberNotFound
				}

				if role.Name == "@everyone" {
					return htmx.AlertUseless
				}

				if role.Name != "@everyone" && request.Name != "" {
					role.Name = request.Name
				}
				if request.Moniker != "" {
					role.Moniker = request.Moniker
				}
				if request.Color != "" {
					role.Color = model.ParseColorHex(request.Color)
				}

				if err := role.Validate(); err != nil {
					return err
				}

				if err := role.Update(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				actionRole := model.ActionRole{
					RoleId:      role.Id,
					CreatorId:   caller.Id,
					GroupId:     group.Id,
					Action:      "update",
					Description: model.Description(fmt.Sprintf("Updated role %s (%s)", role.Moniker, role.Name)),
					ActedAt:     model.TimePast(time.Now()),
				}
				_ = actionRole.Insert()

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/roles/:role_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				roleId := fiber.Params[uint64](ctx, "role_id")

				role, err := model.NewRoleFromId(roleId, group.Id)
				if err != nil || role.IsEmpty() {
					return htmx.AlertGroupMemberNotFound
				}

				if role.Name == "@everyone" {
					return htmx.AlertGroupMemberNotAllowed
				}

				actionRole := model.ActionRole{
					RoleId:      role.Id,
					CreatorId:   caller.Id,
					GroupId:     group.Id,
					Action:      "delete",
					Description: model.Description(fmt.Sprintf("Deleted role %s (%s)", role.Moniker, role.Name)),
					ActedAt:     model.TimePast(time.Now()),
				}
				_ = actionRole.Insert()

				if err := role.Delete(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				_ = model.Delete(model.TableRoleAssignees, goqu.Ex{"role_id": roleId})

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/roles",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			perms.UseBind[RoleAssignForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[RoleAssignForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				userId := fiber.Params[uint64](ctx, "user_id")

				role, err := model.NewRoleFromId(request.RoleId, group.Id)
				if err != nil || role.IsEmpty() {
					return htmx.AlertGroupMemberNotFound
				}

				if model.IsRoleAssigned(userId, role.Id) {
					return htmx.AlertUseless
				}

				roleAssignee := model.RoleAssignee{
					UserId: userId,
					RoleId: role.Id,
				}

				_ = roleAssignee.Insert()

				targetUser, _ := model.NewUserFromId(userId)
				actionRole := model.ActionRole{
					RoleId:      role.Id,
					CreatorId:   caller.Id,
					GroupId:     group.Id,
					Action:      "assign",
					Description: model.Description(fmt.Sprintf("Assigned role %s to %s", role.Moniker, targetUser.Moniker)),
					ActedAt:     model.TimePast(time.Now()),
				}
				_ = actionRole.Insert()

				if userId != caller.Id {
					notif := model.NewNotification(
						userId,
						caller.Id,
						group.Id,
						0,
						0,
						"role_assigned",
						fmt.Sprintf("You were assigned role %s in %s", role.Moniker, group.Moniker),
						fmt.Sprintf("%s assigned you role %s in %s", caller.Moniker, role.Moniker, group.Moniker),
					)
					_ = model.CreateAndPushNotification(app, notif)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/members/:user_id/roles/:role_id",
			perms.RequirePermLevel("group_id", model.PermNameRoles, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				userId := fiber.Params[uint64](ctx, "user_id")
				roleId := fiber.Params[uint64](ctx, "role_id")

				role, _ := model.NewRoleFromId(roleId, group.Id)

				roleAssignee := model.RoleAssignee{
					UserId: userId,
					RoleId: roleId,
				}

				_ = roleAssignee.Delete()

				targetUser, _ := model.NewUserFromId(userId)
				actionRole := model.ActionRole{
					RoleId:      roleId,
					CreatorId:   caller.Id,
					GroupId:     group.Id,
					Action:      "unassign",
					Description: model.Description(fmt.Sprintf("Removed role %s from %s", role.Moniker, targetUser.Moniker)),
					ActedAt:     model.TimePast(time.Now()),
				}
				_ = actionRole.Insert()

				if userId != caller.Id {
					notif := model.NewNotification(
						userId,
						caller.Id,
						group.Id,
						0,
						0,
						"role_unassigned",
						fmt.Sprintf("Role %s was removed in %s", role.Moniker, group.Moniker),
						fmt.Sprintf("%s removed role %s from you in %s", caller.Moniker, role.Moniker, group.Moniker),
					)
					_ = model.CreateAndPushNotification(app, notif)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/roles", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/kick",
			perms.RequirePermLevel("group_id", model.PermNameMembers, model.PermLevelAll),
			perms.UseBind[MemberKickForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MemberKickForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				targetUserId := fiber.Params[uint64](ctx, "user_id")

				if targetUserId == group.OwnerId || targetUserId == caller.Id {
					return htmx.AlertGroupMemberNotAllowed
				}

				targetMember, err := model.NewMemberFromId(group.Id, targetUserId)
				if err != nil || targetMember.IsEmpty() {
					return htmx.AlertGroupMemberNotFound
				}

				if err := targetMember.Kick(caller.Id, request.Description); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				model.RecordUserLog(targetUserId, caller.Id, "kick_group", fmt.Sprintf("Kicked from group %s (%s)", group.Moniker, request.Description))

				var notifContent string
				reasonStr := strings.TrimSpace(string(request.Description))
				if reasonStr != "" {
					notifContent = fmt.Sprintf("Reason: %s", reasonStr)
				}

				notif := model.NewNotification(
					targetUserId,
					caller.Id,
					group.Id,
					0,
					0,
					"kicked",
					fmt.Sprintf("You were kicked from %s", group.Moniker),
					notifContent,
				)
				_ = model.CreateAndPushNotification(app, notif)

				broadcastGroupMemberLeave(app, group, targetUserId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/ban",
			perms.RequirePermLevel("group_id", model.PermNameMembers, model.PermLevelAll),
			perms.UseBind[MemberBanForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MemberBanForm](ctx, perms.LocalForm)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				targetUserId := fiber.Params[uint64](ctx, "user_id")

				if targetUserId == group.OwnerId || targetUserId == caller.Id {
					return htmx.AlertGroupMemberNotAllowed
				}

				durationDays := request.DurationDays
				if durationDays == 0 {
					durationDays = 365
				}
				endsAt := model.TimeFuture(time.Now().AddDate(0, 0, int(durationDays)))

				targetMember, err := model.NewMemberFromId(group.Id, targetUserId)
				if err != nil || targetMember.IsEmpty() {
					targetMember = model.NewMember(group.Id, targetUserId, "")
				}

				banReason := request.Description

				if err := targetMember.Ban(caller.Id, endsAt, banReason); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				model.RecordUserLog(targetUserId, caller.Id, "ban_group", fmt.Sprintf("Banned from group %s (%s)", group.Moniker, banReason))

				var notifContent string
				reasonStr := strings.TrimSpace(string(banReason))
				if reasonStr != "" {
					notifContent = fmt.Sprintf("Reason: %s (Duration: %d days)", reasonStr, durationDays)
				} else {
					notifContent = fmt.Sprintf("Duration: %d days", durationDays)
				}

				notif := model.NewNotification(
					targetUserId,
					caller.Id,
					group.Id,
					0,
					0,
					"banned",
					fmt.Sprintf("You were banned from %s", group.Moniker),
					notifContent,
				)
				_ = model.CreateAndPushNotification(app, notif)

				broadcastGroupMemberLeave(app, group, targetUserId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/unban",
			perms.RequirePermLevel("group_id", model.PermNameMembers, model.PermLevelAll),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				targetUserId := fiber.Params[uint64](ctx, "user_id")

				member := model.NewMember(group.Id, targetUserId, "")
				if err := member.Unban(caller.Id); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				model.RecordUserLog(targetUserId, caller.Id, "unban_group", fmt.Sprintf("Unbanned from group %s", group.Moniker))

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/ws/:group_id/",
			perms.RequirePerm("group_id", model.PermNameRead),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				ctxWs := session.New(ctx)
				ctxWs.GroupId = group.Id
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)

				return websocket.New(func(conn *websocket.Conn) {
					ctxWs.Conn = conn
					session.UserSessionMap.Connect(user.Id, ctxWs)
					broadcastGroupMembersUpdate(app, group)
					defer func() {
						session.UserSessionMap.Close(user.Id, ctxWs)
						broadcastGroupMembersUpdate(app, group)
					}()
					for !ctxWs.Closed {
						messageType, message, err := ctxWs.Conn.ReadMessage()
						if err != nil {
							break
						}

						start := time.Now()
						ctxWs.MessageType = messageType
						ctxWs.Message = message
						err = ctxWs.Flush()

						logWS(start, err, ctxWs)

						if err != nil {
							break
						}
					}
					ctxWs.Closed = true
					_ = ctxWs.Conn.Close()
				})(ctx)
			},
		)
}

func broadcastGroupMembersUpdate(app *fiber.App, group model.Group) {
	const MembersPagination uint = 5
	memberList := group.UsersPage(1, MembersPagination)
	buf, err := render.RenderBuffer(app, "partials/chat-members", fiber.Map{
		"Group":             group,
		"MemberList":        memberList,
		"MembersPage":       1,
		"MembersPagination": MembersPagination,
	})
	if err != nil {
		return
	}
	str := buf.String()
	membersOob := render.WrapOob("innerHTML:#members", &str)

	countStr := fmt.Sprintf("%d", group.MembersCount())
	countOob := render.WrapOob("innerHTML:#members-count", &countStr)

	session.UserSessionMap.PushGroup(group.Id, membersOob+countOob)
}

func broadcastGroupMemberLeave(app *fiber.App, group model.Group, userId uint64) {
	deleteOob := render.WrapOob(fmt.Sprintf("delete:#member-%d", userId), nil)

	redirectUrl := "/chat"
	if group.IsBanned(userId) {
		redirectUrl = fmt.Sprintf("/chat/groups/join/%s", group.Name)
	}
	redirectOob := fmt.Sprintf(`<div id="redirect-kicked" hx-swap-oob="beforeend:body"><script>if(window.location.pathname==='/chat/groups/%d'||window.location.pathname.startsWith('/chat/groups/%d/')){window.location.href='%s';}</script></div>`, group.Id, group.Id, redirectUrl)
	session.UserSessionMap.PushUser(userId, redirectOob)

	const MembersPagination uint = 5
	memberList := group.UsersPage(1, MembersPagination)
	buf, err := render.RenderBuffer(app, "partials/chat-members", fiber.Map{
		"Group":             group,
		"MemberList":        memberList,
		"MembersPage":       1,
		"MembersPagination": MembersPagination,
	})
	if err != nil {
		session.UserSessionMap.PushGroup(group.Id, deleteOob)
		return
	}
	str := buf.String()
	membersOob := render.WrapOob("innerHTML:#members", &str)

	countStr := fmt.Sprintf("%d", group.MembersCount())
	countOob := render.WrapOob("innerHTML:#members-count", &countStr)

	session.UserSessionMap.PushGroup(group.Id, deleteOob+membersOob+countOob)
}

func handleMemberSearch(ctx fiber.Ctx) error {
	group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
	request := fiber.Locals[MemberSearchForm](ctx, perms.LocalForm)

	results, err := group.SearchUsers(request.Q, 10)
	if err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	var filtered []model.User
	for _, u := range results {
		if request.RoleId != 0 && model.IsRoleAssigned(u.Id, request.RoleId) {
			continue
		}
		if request.ChannelId != 0 && model.IsChannelAssignee(request.ChannelId, u.Id) {
			continue
		}
		filtered = append(filtered, u)
	}

	if htmx.IsHtmx(ctx) {
		return ctx.Render("partials/member-search-results", fiber.Map{
			"Group":     group,
			"Results":   filtered,
			"ChannelId": request.ChannelId,
			"RoleId":    request.RoleId,
		})
	}

	return ctx.JSON(filtered)
}

type EntityResultItem struct {
	Id      uint64
	Moniker string
	Name    string
	Subtext string
}

func handleEntitySearch(ctx fiber.Ctx) error {
	group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
	request := fiber.Locals[EntitySearchForm](ctx, perms.LocalForm)

	perm, _ := model.NewPermissionFromId(request.PermId, group.Id)

	var items []EntityResultItem
	q := strings.ToLower(strings.TrimSpace(request.Q))

	switch request.EntityType {
	case "channel":
		attachedMap := make(map[uint64]bool)
		for _, ch := range perm.Channels {
			attachedMap[ch.Id] = true
		}
		channels, _ := model.GetGroupChannels(group.Id)
		for _, ch := range channels {
			if attachedMap[ch.Id] {
				continue
			}
			if q == "" || strings.Contains(strings.ToLower(string(ch.Moniker)), q) || strings.Contains(strings.ToLower(string(ch.Name)), q) {
				items = append(items, EntityResultItem{
					Id:      ch.Id,
					Moniker: string(ch.Moniker),
					Name:    string(ch.Name),
					Subtext: string(ch.Type),
				})
			}
		}
	case "role":
		attachedMap := make(map[uint64]bool)
		for _, r := range perm.Roles {
			attachedMap[r.Id] = true
		}
		roles := group.Roles()
		for _, r := range roles {
			if attachedMap[r.Id] {
				continue
			}
			if q == "" || strings.Contains(strings.ToLower(string(r.Moniker)), q) || strings.Contains(strings.ToLower(string(r.Name)), q) {
				items = append(items, EntityResultItem{
					Id:      r.Id,
					Moniker: string(r.Moniker),
					Name:    string(r.Name),
					Subtext: "Role",
				})
			}
		}
	case "user":
		attachedMap := make(map[uint64]bool)
		for _, u := range perm.Users {
			attachedMap[u.Id] = true
		}
		users, _ := group.SearchUsers(request.Q, 10)
		for _, u := range users {
			if attachedMap[u.Id] {
				continue
			}
			items = append(items, EntityResultItem{
				Id:      u.Id,
				Moniker: string(u.Moniker),
				Name:    string(u.Name),
				Subtext: "User",
			})
		}
	}

	if htmx.IsHtmx(ctx) {
		return ctx.Render("partials/perm-entity-search-results", fiber.Map{
			"Group":      group,
			"PermId":     request.PermId,
			"EntityType": request.EntityType,
			"Results":    items,
		})
	}

	return ctx.JSON(items)
}

func handleTicketSearch(ctx fiber.Ctx) error {
	group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
	request := fiber.Locals[TicketSearchForm](ctx, perms.LocalForm)
	mode := request.Mode
	if mode == "" {
		mode = "parent"
	}

	results, err := model.SearchTicketChannels(group.Id, request.Q, request.ExcludeId, 10)
	if err != nil {
		return htmx.AlertDatabase.Join(err)
	}

	if request.ExcludeId != 0 && mode == "parent" {
		existingParents, _ := model.GetParentChannels(request.ExcludeId)
		if len(existingParents) > 0 {
			parentMap := make(map[uint64]bool)
			for _, p := range existingParents {
				parentMap[p.Id] = true
			}
			var filtered []model.Channel
			for _, ch := range results {
				if !parentMap[ch.Id] {
					filtered = append(filtered, ch)
				}
			}
			results = filtered
		}
	}

	if mode == "duplicate" {
		var filtered []model.Channel
		for _, ch := range results {
			if len(ch.ParentChannels()) > 0 {
				filtered = append(filtered, ch)
			}
		}
		results = filtered
	}

	if htmx.IsHtmx(ctx) {
		return ctx.Render("partials/ticket-search-results", fiber.Map{
			"Results": results,
			"Mode":    mode,
		})
	}

	return ctx.JSON(results)
}

func saveMessageAttachment(ctx fiber.Ctx, groupId uint64, message *model.Message) error {
	file, errFile := ctx.FormFile("attachment")
	if errFile != nil {
		file, errFile = ctx.FormFile("file")
	}
	if errFile == nil && file != nil {
		if err := os.MkdirAll("uploads", 0755); err != nil {
			return htmx.AlertDatabase.Join(err)
		}
		ext := filepath.Ext(file.Filename)
		sanitizedName := filepath.Base(file.Filename)
		filename := fmt.Sprintf("file-%d-%d%s", groupId, time.Now().UnixNano(), ext)
		dst := filepath.Join("uploads", filename)
		if err := ctx.SaveFile(file, dst); err != nil {
			return htmx.AlertDatabase.Join(err)
		}
		message.AttachmentUrl = "/uploads/" + filename
		message.AttachmentName = sanitizedName
		message.AttachmentSize = file.Size
		message.AttachmentType = file.Header.Get("Content-Type")
	}
	return nil
}
