package model

import (
	"strings"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type PermSwitch string

const (
	PermSwitchNone     PermSwitch = ""
	PermSwitchDisallow PermSwitch = "disallow"
	PermSwitchAllow    PermSwitch = "allow"
)

func (perm PermSwitch) IsValid() bool {
	return perm == PermSwitchNone ||
		perm == PermSwitchDisallow ||
		perm == PermSwitchAllow
}

func (perm PermSwitch) Has() bool {
	return perm == PermSwitchAllow
}

type PermMessages string

const (
	PermMessagesNone   PermMessages = ""
	PermMessagesHidden PermMessages = "hidden" // Can not see any messages. Channel info can be available.
	PermMessagesRead   PermMessages = "read"   // Can read chat messages.
	PermMessagesWrite  PermMessages = "write"  // Can read, write and delete own messages.
	PermMessagesDelete PermMessages = "delete" // Can read, write and delete own messages. Can delete other people's messages.
)

func (perm PermMessages) IsValid() bool {
	return perm == PermMessagesNone ||
		perm == PermMessagesHidden ||
		perm == PermMessagesRead ||
		perm == PermMessagesWrite ||
		perm == PermMessagesDelete
}

func (perm PermMessages) CanReadMessages() bool {
	return perm == PermMessagesRead ||
		perm == PermMessagesWrite ||
		perm == PermMessagesDelete
}

func (perm PermMessages) CanWriteMessages() bool {
	return perm == PermMessagesWrite ||
		perm == PermMessagesDelete
}

func (perm PermMessages) CanDeleteMessages() bool {
	return perm == PermMessagesDelete
}

type PermMembers string

const (
	PermMembersNone   PermMembers = ""
	PermMembersRead   PermMembers = "read"   // Can see all information about all users.
	PermMembersInvite PermMembers = "invite" // Can invite new users.
	PermMembersWrite  PermMembers = "write"  // Can invite new users and change other people's monikers.
	PermMembersDelete PermMembers = "delete" // Can invite, change monikers, kick and ban people.
)

func (perm PermMembers) IsValid() bool {
	return perm == PermMembersNone ||
		perm == PermMembersRead ||
		perm == PermMembersInvite ||
		perm == PermMembersWrite ||
		perm == PermMembersDelete
}

func (perm PermMembers) CanSee() bool {
	return perm == PermMembersRead ||
		perm == PermMembersInvite ||
		perm == PermMembersWrite ||
		perm == PermMembersDelete
}

func (perm PermMembers) CanInvite() bool {
	return perm == PermMembersInvite ||
		perm == PermMembersWrite ||
		perm == PermMembersDelete
}

func (perm PermMembers) CanManage() bool {
	return perm == PermMembersWrite ||
		perm == PermMembersDelete
}

func (perm PermMembers) CanKickBan() bool {
	return perm == PermMembersDelete
}

type Role struct {
	Id      uint32  `db:"id"`
	GroupId uint64  `db:"group_id"`
	Name    Name    `db:"name"`
	Moniker Moniker `db:"moniker"`
	Color   Color   `db:"color"`

	PermMessages    PermMessages `db:"perm_messages"`
	PermRoles       PermSwitch   `db:"perm_roles"`
	PermMembers     PermMembers  `db:"perm_members"`
	PermGroupChange PermSwitch   `db:"perm_group_change"`
	PermAdmin       PermSwitch   `db:"perm_admin"`
}

var _ Model = (*Role)(nil)

func (role Role) Validate() htmx.Alert {
	if !role.Name.IsValid() {
		return htmx.AlertFormatName
	}
	if !role.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !role.Color.IsValid() {
		return htmx.AlertFormatColor
	}

	if !role.PermMessages.IsValid() {
		return htmx.AlertFormatGroupPermMessages
	}
	if !role.PermRoles.IsValid() {
		return htmx.AlertFormatGroupPermSwitch
	}
	if !role.PermMembers.IsValid() {
		return htmx.AlertFormatGroupPermMembers
	}
	if !role.PermGroupChange.IsValid() {
		return htmx.AlertFormatGroupPermSwitch
	}
	if !role.PermAdmin.IsValid() {
		return htmx.AlertFormatGroupPermSwitch
	}

	return nil
}

func (role Role) IsEmpty() bool {
	return role.Id == 0 || role.Name == ""
}

// permissions
//
// NOTE: Keep 'none' at the end and 'disallow' at the first places.
var (
	permSwitch   = []PermSwitch{PermSwitchDisallow, PermSwitchAllow, PermSwitchNone}
	permMembers  = []PermMembers{PermMembersRead, PermMembersInvite, PermMembersDelete, PermMembersNone}
	permMessages = []PermMessages{PermMessagesHidden, PermMessagesRead, PermMessagesWrite, PermMessagesDelete, PermMessagesNone}
)

// Enabled rights have priority over disabled rights.
func (r *Role) Merge(roleList ...Role) {
	for _, role := range roleList {
		r.PermMessages = mergePerm(permMessages, r.PermMessages, role.PermMessages)
		r.PermRoles = mergePerm(permSwitch, r.PermRoles, role.PermRoles)
		r.PermMembers = mergePerm(permMembers, r.PermMembers, role.PermMembers)
		r.PermGroupChange = mergePerm(permSwitch, r.PermGroupChange, role.PermGroupChange)
		r.PermAdmin = mergePerm(permSwitch, r.PermAdmin, role.PermAdmin)
	}
}

// Enabled rights have priority over disabled rights.
func mergePerm[T PermSwitch | PermMessages | PermMembers](list []T, perm1, perm2 T) T {
	for _, perm := range list {
		if perm1 == perm || perm2 == perm {
			return perm
		}
	}
	listStr := make([]string, len(list))
	for i, v := range list {
		listStr[i] = string(v)
	}
	panic("unexpected perm msg value: " + string(perm1) + " or " + string(perm2) + ". available values: " + strings.Join(listStr, ",") + ".")
}

func NewRoleFromId(id uint32, groupId uint64) (Role, error) {
	role := Role{}
	return role, First(TableRoles, goqu.Ex{"id": id, "group_id": groupId}, &role)
}

func NewRoleFromName(name Name, groupId uint64) (Role, error) {
	role := Role{}
	return role, First(TableRoles, goqu.Ex{"name": name, "group_id": groupId}, &role)
}

const roleNameEveryone Name = "@everyone"

func NewRoleEveryone(groupId uint64) Role {
	return Role{
		GroupId: groupId,
		Name:    roleNameEveryone,
		Moniker: "everyone",

		PermMessages:    PermMessagesWrite,
		PermRoles:       PermSwitchDisallow,
		PermMembers:     PermMembersRead,
		PermGroupChange: PermSwitchDisallow,
		PermAdmin:       PermSwitchDisallow,
	}
}

func NewAllAccessRole(allow bool, role Role) Role {
	if allow {
		role.PermMessages = PermMessagesDelete
		role.PermRoles = PermSwitchAllow
		role.PermMembers = PermMembersDelete
		role.PermGroupChange = PermSwitchAllow
		role.PermAdmin = PermSwitchAllow
		return role
	}

	role.PermMessages = PermMessagesHidden
	role.PermRoles = PermSwitchDisallow
	role.PermMembers = PermMembersRead
	role.PermGroupChange = PermSwitchDisallow
	role.PermAdmin = PermSwitchDisallow
	return role
}

func (role *Role) Insert() error {
	return InsertId(TableRoles, role, &role.Id)
}

func (role Role) Update() error {
	return Update(TableRoles, role, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}

func (role Role) Delete() error {
	return Delete(TableRoles, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}
