package routes

import (
	"fmt"
	"os"
	"path/filepath"
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

type RoleCreate struct {
	Name            model.Name         `form:"name"`
	Moniker         model.Moniker      `form:"moniker"`
	Color           string             `form:"color"`
	PermMessages    model.PermMessages `form:"perm_messages"`
	PermRoles       model.PermSwitch   `form:"perm_roles"`
	PermMembers     model.PermMembers  `form:"perm_members"`
	PermGroupChange model.PermSwitch   `form:"perm_group_change"`
	PermAdmin       model.PermSwitch   `form:"perm_admin"`
}

type RoleChange struct {
	Name            model.Name         `form:"name"`
	Moniker         model.Moniker      `form:"moniker"`
	Color           string             `form:"color"`
	PermMessages    model.PermMessages `form:"perm_messages"`
	PermRoles       model.PermSwitch   `form:"perm_roles"`
	PermMembers     model.PermMembers  `form:"perm_members"`
	PermGroupChange model.PermSwitch   `form:"perm_group_change"`
	PermAdmin       model.PermSwitch   `form:"perm_admin"`
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

type ChannelPermissionForm struct {
	RoleId          uint64             `form:"role_id"`
	PermMessages    model.PermMessages `form:"perm_messages"`
	PermRoles       model.PermSwitch   `form:"perm_roles"`
	PermMembers     model.PermMembers  `form:"perm_members"`
	PermGroupChange model.PermSwitch   `form:"perm_group_change"`
	PermAdmin       model.PermSwitch   `form:"perm_admin"`
	Inherit         bool               `form:"inherit"`
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
		Post("/:group_id/roles/:role_id/move",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				roleId := fiber.Params[uint64](ctx, "role_id")
				direction := ctx.Query("direction", "up")

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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
		Post("/:group_id/channels/create",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
		Post("/:group_id/channels/:channel_id/permissions",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
			perms.UseBind[ChannelPermissionForm](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[ChannelPermissionForm](ctx, perms.LocalForm)
				channelId := fiber.Params[uint64](ctx, "channel_id")
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				cp := model.ChannelPermission{
					ChannelId:       channelId,
					RoleId:          request.RoleId,
					PermMessages:    request.PermMessages,
					PermRoles:       request.PermRoles,
					PermMembers:     request.PermMembers,
					PermGroupChange: request.PermGroupChange,
					PermAdmin:       request.PermAdmin,
					Inherit:         request.Inherit,
				}

				if err := model.SetChannelPermission(cp); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/channels", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/channels/:channel_id/messages/create",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				channelId := fiber.Params[uint64](ctx, "channel_id")
				if channelId != 0 {
					role = model.EffectiveRolePermission(channelId, role)
				}
				return role.PermMessages.CanWriteMessages()
			}),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				channelId := fiber.Params[uint64](ctx, "channel_id")

				message := model.NewMessageFilled(channelId, group.Id, user.Id, request.Content)
				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()

					session.UserSessionMap.Push(
						render.WrapOob("beforeend:#chat", &str),
						session.PickMessages,
					)

					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/messages/create",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanWriteMessages()
			}),
			perms.UseBind[MessageCreate](),
			func(ctx fiber.Ctx) error {
				request := fiber.Locals[MessageCreate](ctx, perms.LocalForm)
				user := fiber.Locals[model.User](ctx, perms.LocalAuth)
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)

				channelId := request.ChannelId
				if channelId == 0 {
					channels, _ := model.GetGroupChannels(group.Id)
					if len(channels) > 0 {
						channelId = channels[0].Id
					} else {
						return htmx.AlertGroupNotFound
					}
				}

				message := model.NewMessageFilled(channelId, group.Id, user.Id, request.Content)
				if err := message.Validate(); err != nil {
					return err
				}

				if err := message.Insert(); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					buf, err := render.RenderBuffer(app, "partials/chat-messages", fiber.Map{
						"MessageList": []model.Message{message},
					})
					if err != nil {
						return err
					}
					str := buf.String()

					session.UserSessionMap.Push(
						render.WrapOob("beforeend:#chat", &str),
						session.PickMessages,
					)

					return nil
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
		Get("/:group_id/channels/:channel_id/messages/page/:messages_page",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				channelId := fiber.Params[uint64](ctx, "channel_id")
				if channelId != 0 {
					role = model.EffectiveRolePermission(channelId, role)
				}
				return role.PermMessages.CanReadMessages()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				page := fiber.Params[uint](ctx, "messages_page")
				const MessagesPagination uint = 5

				channels, _ := model.GetGroupChannels(group.Id)
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMembers.CanSee()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
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
					GroupId:         group.Id,
					Name:            request.Name,
					Moniker:         request.Moniker,
					Color:           model.ParseColorHex(request.Color),
					PermMessages:    request.PermMessages,
					PermRoles:       request.PermRoles,
					PermMembers:     request.PermMembers,
					PermGroupChange: request.PermGroupChange,
					PermAdmin:       request.PermAdmin,
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
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

				if role.Name != "@everyone" && request.Name != "" {
					role.Name = request.Name
				}
				if request.Moniker != "" {
					role.Moniker = request.Moniker
				}
				if request.Color != "" {
					role.Color = model.ParseColorHex(request.Color)
				}
				if request.PermMessages != "" {
					role.PermMessages = request.PermMessages
				}
				if request.PermRoles != "" {
					role.PermRoles = request.PermRoles
				}
				if request.PermMembers != "" {
					role.PermMembers = request.PermMembers
				}
				if request.PermGroupChange != "" {
					role.PermGroupChange = request.PermGroupChange
				}
				if request.PermAdmin != "" {
					role.PermAdmin = request.PermAdmin
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
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

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Delete("/:group_id/members/:user_id/roles/:role_id",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has()
			}),
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

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/kick",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMembers.CanKickBan()
			}),
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

				broadcastGroupMemberLeave(app, group, targetUserId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/ban",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMembers.CanKickBan()
			}),
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

				if err := targetMember.Ban(caller.Id, endsAt, request.Description); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				broadcastGroupMemberLeave(app, group, targetUserId)

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Post("/:group_id/members/:user_id/unban",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMembers.CanKickBan()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				caller := fiber.Locals[model.User](ctx, perms.LocalAuth)
				targetUserId := fiber.Params[uint64](ctx, "user_id")

				member := model.NewMember(group.Id, targetUserId, "")
				if err := member.Unban(caller.Id); err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					htmx.Redirect(ctx, fmt.Sprintf("/groups/%d/settings/members", group.Id))
					return nil
				}

				return ctx.SendStatus(fiber.StatusOK)
			},
		).
		Get("/ws/:group_id/",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
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
