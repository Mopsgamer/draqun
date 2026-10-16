package routes

import (
	"fmt"
	"html/template"
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
		Post("/:group_id/channels/:channel_id/move",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermGroupChange.Has()
			}),
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
		Get("/:group_id/members/search",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermRoles.Has() || role.PermMembers.CanSee() || role.PermMessages.CanReadMessages()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				q := ctx.Query("q")
				roleId := fiber.Query[uint64](ctx, "role_id")
				channelId := fiber.Query[uint64](ctx, "channel_id")

				results, err := group.SearchUsers(q, 10)
				if err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					if len(results) == 0 {
						return ctx.Type("html").SendString(`<div class="p-2 text-xs text-base-content/50 italic text-center">No matching members found</div>`)
					}
					var htmlStr string
					for _, u := range results {
						monikerEscaped := template.HTMLEscapeString(string(u.Moniker))
						if channelId != 0 {
							htmlStr += fmt.Sprintf(`
							<form hx-post="/groups/%d/channels/%d/assignees" class="w-full text-left flex items-center justify-between py-1.5 hover:bg-surface-card-hover text-xs transition-colors cursor-pointer">
								<input type="hidden" name="user_id" value="%d" />
								<div class="flex items-center gap-2 truncate min-w-0">
									<div class="avatar shrink-0"><div class="w-4 h-4 rounded-full"><img src="%s" class="size-full object-cover rounded-full" alt="Avatar"></div></div>
									<span class="truncate font-medium">%s</span>
									<span class="font-mono text-base-content/40 text-[11px] truncate">@%s</span>
								</div>
								<button type="submit" class="btn btn-primary btn-xs shrink-0 cursor-pointer">Assign</button>
							</form>`, group.Id, channelId, u.Id, u.AvatarUrl(), monikerEscaped, u.Name)
						} else {
							htmlStr += fmt.Sprintf(`
							<form hx-post="/groups/%d/members/%d/roles" class="w-full text-left flex items-center justify-between py-1.5 hover:bg-surface-card-hover text-xs transition-colors cursor-pointer">
								<input type="hidden" name="role_id" value="%d" />
								<div class="flex items-center gap-2 truncate min-w-0">
									<div class="avatar shrink-0"><div class="w-4 h-4 rounded-full"><img src="%s" class="size-full object-cover rounded-full" alt="Avatar"></div></div>
									<span class="truncate font-medium">%s</span>
									<span class="font-mono text-base-content/40 text-[11px] truncate">@%s</span>
								</div>
								<button type="submit" class="btn btn-primary btn-xs shrink-0 cursor-pointer">Assign</button>
							</form>`, group.Id, u.Id, roleId, u.AvatarUrl(), monikerEscaped, u.Name)
						}
					}
					return ctx.Type("html").SendString(htmlStr)
				}

				return ctx.JSON(results)
			},
		).
		Post("/:group_id/channels/:channel_id/assignees",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanWriteMessages()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanWriteMessages()
			}),
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				return role.PermMessages.CanReadMessages()
			}),
			func(ctx fiber.Ctx) error {
				group := fiber.Locals[model.Group](ctx, perms.LocalGroup)
				q := ctx.Query("q")
				excludeId := fiber.Query[uint64](ctx, "exclude_id")
				mode := ctx.Query("mode", "parent")

				results, err := model.SearchTicketChannels(group.Id, q, excludeId, 10)
				if err != nil {
					return htmx.AlertDatabase.Join(err)
				}

				if htmx.IsHtmx(ctx) {
					if len(results) == 0 {
						return ctx.Type("html").SendString(`<div class="p-2 text-xs text-base-content/50 italic text-center">No matching tickets found</div>`)
					}
					var html string
					for _, ch := range results {
						iconClass := "icon-[lucide--circle-dot] text-emerald-500"
						if ch.Status == "closed" {
							iconClass = "icon-[lucide--check-circle-2] text-purple-500"
						}
						monikerEscaped := template.HTMLEscapeString(string(ch.Moniker))
						jsTitleEscaped := template.JSEscapeString(string(ch.Moniker))

						if mode == "duplicate" {
							html += fmt.Sprintf(`
							<button type="button" class="w-full text-left flex items-center justify-between py-1.5 hover:bg-surface-card-hover text-xs transition-colors cursor-pointer" onclick="var inp=document.getElementById('duplicate_of_id_input'); if(inp){inp.value='%d';} var prev=document.getElementById('duplicate_selected_preview'); if(prev){prev.innerHTML='<span class=\'inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-amber-500/15 text-amber-500 text-xs font-medium\'><span class=\'icon icon-[lucide--copy-x] size-3\'></span>#%d %s</span>';}">
								<div class="flex items-center gap-2 truncate min-w-0">
									<span class="icon %s size-3.5 shrink-0"></span>
									<span class="truncate font-medium">%s</span>
								</div>
								<span class="font-mono text-base-content/40 text-[11px] shrink-0 ml-2">#%d</span>
							</button>`, ch.Id, ch.Id, jsTitleEscaped, iconClass, monikerEscaped, ch.Id)
						} else {
							html += fmt.Sprintf(`
							<button type="button" class="w-full text-left flex items-center justify-between py-1.5 hover:bg-surface-card-hover text-xs transition-colors cursor-pointer" onclick="var b=document.getElementById('active-parent-badges'); if(b){ var span=document.createElement('span'); span.className='badge badge-primary badge-sm gap-1 font-mono items-center'; span.innerHTML='<input type=\'hidden\' name=\'parent_ids\' value=\'%d\' /><span>#%d %s</span><button type=\'button\' class=\'btn btn-ghost btn-xs btn-circle p-0 h-auto min-h-0 text-white/80 hover:text-white\' onclick=\'this.parentElement.remove()\'>✕</button>'; b.appendChild(span); }">
								<div class="flex items-center gap-2 truncate min-w-0">
									<span class="icon %s size-3.5 shrink-0"></span>
									<span class="truncate font-medium">%s</span>
								</div>
								<span class="font-mono text-base-content/40 text-[11px] shrink-0 ml-2">#%d</span>
							</button>`, ch.Id, ch.Id, jsTitleEscaped, iconClass, monikerEscaped, ch.Id)
						}
					}
					return ctx.Type("html").SendString(html)
				}

				return ctx.JSON(results)
			},
		).
		Post("/:group_id/channels/:parent_id/tickets/create",
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				parentId := fiber.Params[uint64](ctx, "parent_id")
				if parentId != 0 {
					role = model.EffectiveRolePermission(parentId, role)
				}
				return role.PermMessages.CanWriteMessages()
			}),
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

				nameStr := fmt.Sprintf("ticket_%d", time.Now().UnixNano())
				channel := model.NewChannel(model.Name(nameStr), model.Moniker(request.Title), model.Description(request.Content), model.ChannelTypeTicket)
				channel.AuthorId = user.Id
				channel.Status = "open"

				if err := channel.Validate(); err != nil {
					return err
				}

				if err := channel.Insert(); err != nil {
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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				channelId := fiber.Params[uint64](ctx, "channel_id")
				if channelId != 0 {
					role = model.EffectiveRolePermission(channelId, role)
				}
				return role.PermMessages.CanWriteMessages()
			}),
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
				_ = model.SetParentChannels(channel.Id, request.ParentIds)

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
			perms.MemberByAuthAndGroupId("group_id", func(ctx fiber.Ctx, role model.Role) bool {
				channelId := fiber.Params[uint64](ctx, "channel_id")
				if channelId != 0 {
					role = model.EffectiveRolePermission(channelId, role)
				}
				return role.PermMessages.CanWriteMessages()
			}),
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

				notif := model.NewNotification(
					targetUserId,
					caller.Id,
					group.Id,
					0,
					0,
					"kicked",
					fmt.Sprintf("You were kicked from %s", group.Moniker),
					fmt.Sprintf("Reason: %s", request.Description),
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

				notif := model.NewNotification(
					targetUserId,
					caller.Id,
					group.Id,
					0,
					0,
					"banned",
					fmt.Sprintf("You were banned from %s", group.Moniker),
					fmt.Sprintf("Reason: %s", request.Description),
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
