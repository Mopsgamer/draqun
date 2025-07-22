package database

import (
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type PermSwitch string

const (
	PermSwitchNone     PermSwitch = ""
	PermSwitchDisallow PermSwitch = "disallow"
	PermSwitchAllow    PermSwitch = "allow"
)

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
	PermMembersWrite  PermMembers = "write"  // Can invite new users and change other people's nicknames.
	PermMembersDelete PermMembers = "delete" // Can invite, change nicknames, kick and ban people.
)

type Role struct {
	Db *goqu.Database `db:"-"`

	Id      uint32 `db:"id"`
	GroupId uint64 `db:"group_id"`
	Name    string `db:"name"`
	Moniker string `db:"moniker"`
	Color   uint32 `db:"color"`

	PermMessages    PermMessages `db:"perm_messages"`
	PermRoles       PermSwitch   `db:"perm_roles"`
	PermMembers     PermMembers  `db:"perm_members"`
	PermGroupChange PermSwitch   `db:"perm_group_change"`
	PermAdmin       PermSwitch   `db:"perm_admin"`
}

func (role Role) isEmpty() bool {
	return role.Id == 0 && role.Name == ""
}

// permissions
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

func NewRole(db *goqu.Database) Role {
	return Role{Db: db}
}

func NewRoleEveryone(db *goqu.Database, groupId uint64) Role {
	return Role{
		Db:      db,
		GroupId: groupId,
		Name:    "@everyone",
		Moniker: "everyone",

		PermMessages:    PermMessagesRead,
		PermRoles:       PermSwitchDisallow,
		PermMembers:     PermMembersRead,
		PermGroupChange: PermSwitchDisallow,
		PermAdmin:       PermSwitchDisallow,
	}
}

func (role *Role) Insert() bool {
	id := Insert(role.Db, TableRoles, role)
	role.Id = uint32(id)
	return id != 0
}

func (role *Role) Update() bool {
	return Update(role.Db, TableRoles, role, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}

func (role *Role) Delete() bool {
	return Delete(role.Db, TableRoles, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}

func (role *Role) FromId(id uint32, groupId uint64) bool {
	First(role.Db, TableRoles, goqu.Ex{"id": id, "group_id": groupId}, role)
	return role.isEmpty()
}

func (role *Role) FromName(name string, groupId uint64) bool {
	First(role.Db, TableRoles, goqu.Ex{"name": name, "group_id": groupId}, role)
	return role.isEmpty()
}
