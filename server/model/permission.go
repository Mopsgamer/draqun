package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type PermName string

const (
	PermNameRead     PermName = "read"
	PermNameWrite    PermName = "write"
	PermNameDelete   PermName = "delete"
	PermNameMembers  PermName = "members"
	PermNameRoles    PermName = "roles"
	PermNameSettings PermName = "settings"
	PermNameAdmin    PermName = "admin"
)

func (p PermName) IsValid() bool {
	switch p {
	case PermNameRead, PermNameWrite, PermNameDelete, PermNameMembers, PermNameRoles, PermNameSettings, PermNameAdmin:
		return true
	default:
		return false
	}
}

type PermLevel string

const (
	PermLevelAll     PermLevel = "all"
	PermLevelOwn     PermLevel = "own"
	PermLevelDisable PermLevel = "disable"
)

func (p PermLevel) IsValid() bool {
	switch p {
	case PermLevelAll, PermLevelOwn, PermLevelDisable:
		return true
	default:
		return false
	}
}

type Permission struct {
	checkEmpty
	Id       uint64    `db:"id" goqu:"skipinsert"`
	GroupId  uint64    `db:"group_id"`
	Name     PermName  `db:"name"`
	Level    PermLevel `db:"level"`
	Priority int       `db:"priority"`

	Channels []Channel `db:"-" goqu:"skipinsert"`
	Roles    []Role    `db:"-" goqu:"skipinsert"`
	Users    []User    `db:"-" goqu:"skipinsert"`
}

var _ Model = (*Permission)(nil)

func (p Permission) Validate() htmx.Alert {
	if p.GroupId == 0 {
		return htmx.AlertGroupNotFound
	}
	if !p.Name.IsValid() {
		return htmx.AlertUseless
	}
	if !p.Level.IsValid() {
		return htmx.AlertUseless
	}
	return nil
}

func (p *Permission) Insert() error {
	return InsertId(TablePermissions, p, &p.Id)
}

func (p Permission) Update() error {
	return Update(TablePermissions, p, goqu.Ex{"id": p.Id, "group_id": p.GroupId})
}

func (p Permission) Delete() error {
	_ = Delete(TablePermChannels, goqu.Ex{"perm_id": p.Id})
	_ = Delete(TablePermUsers, goqu.Ex{"perm_id": p.Id})
	_ = Delete(TablePermRoles, goqu.Ex{"perm_id": p.Id})
	return Delete(TablePermissions, goqu.Ex{"id": p.Id, "group_id": p.GroupId})
}

func NewPermissionFromId(id, groupId uint64) (Permission, error) {
	p := Permission{}
	err := First(TablePermissions, goqu.Ex{"id": id, "group_id": groupId}, &p)
	if err == nil {
		_ = p.LoadEntities()
	}
	return p, err
}

func (p *Permission) LoadEntities() error {
	p.Channels = []Channel{}
	sqlCh, argsCh, errCh := Goqu.From(TableChannels).
		Select(goqu.I("app_channels.*")).
		Join(
			goqu.T(TablePermChannels),
			goqu.On(goqu.I("app_channels.id").Eq(goqu.I("app_perm_channels.channel_id"))),
		).
		Where(goqu.I("app_perm_channels.perm_id").Eq(p.Id)).
		Prepared(true).
		ToSQL()
	if errCh == nil {
		var channels []Channel
		if err := Sqlx.Select(&channels, sqlCh, argsCh...); err == nil {
			for i := range channels {
				channels[i].SetFullfilled()
			}
			p.Channels = channels
		}
	}

	p.Roles = []Role{}
	sqlRo, argsRo, errRo := Goqu.From(TableRoles).
		Select(goqu.I("app_group_roles.*")).
		Join(
			goqu.T(TablePermRoles),
			goqu.On(goqu.I("app_group_roles.id").Eq(goqu.I("app_perm_roles.role_id"))),
		).
		Where(goqu.I("app_perm_roles.perm_id").Eq(p.Id)).
		Prepared(true).
		ToSQL()
	if errRo == nil {
		var roles []Role
		if err := Sqlx.Select(&roles, sqlRo, argsRo...); err == nil {
			for i := range roles {
				roles[i].SetFullfilled()
			}
			p.Roles = roles
		}
	}

	p.Users = []User{}
	sqlUs, argsUs, errUs := Goqu.From(TableUsers).
		Select(goqu.I("app_users.*")).
		Join(
			goqu.T(TablePermUsers),
			goqu.On(goqu.I("app_users.id").Eq(goqu.I("app_perm_users.user_id"))),
		).
		Where(goqu.I("app_perm_users.perm_id").Eq(p.Id)).
		Prepared(true).
		ToSQL()
	if errUs == nil {
		var users []User
		if err := Sqlx.Select(&users, sqlUs, argsUs...); err == nil {
			for i := range users {
				users[i].SetFullfilled()
			}
			p.Users = users
		}
	}

	return nil
}

func (p *Permission) AddChannel(channelId uint64) error {
	type PermChannel struct {
		ChannelId uint64 `db:"channel_id"`
		PermId    uint64 `db:"perm_id"`
	}
	pc := PermChannel{ChannelId: channelId, PermId: p.Id}
	return Insert(TablePermChannels, &pc)
}

func (p *Permission) RemoveChannel(channelId uint64) error {
	return Delete(TablePermChannels, goqu.Ex{"perm_id": p.Id, "channel_id": channelId})
}

func (p *Permission) AddRole(roleId uint64) error {
	type PermRole struct {
		RoleId uint64 `db:"role_id"`
		PermId uint64 `db:"perm_id"`
	}
	pr := PermRole{RoleId: roleId, PermId: p.Id}
	return Insert(TablePermRoles, &pr)
}

func (p *Permission) RemoveRole(roleId uint64) error {
	return Delete(TablePermRoles, goqu.Ex{"perm_id": p.Id, "role_id": roleId})
}

func (p *Permission) AddUser(userId uint64) error {
	type PermUser struct {
		UserId uint64 `db:"user_id"`
		PermId uint64 `db:"perm_id"`
	}
	pu := PermUser{UserId: userId, PermId: p.Id}
	return Insert(TablePermUsers, &pu)
}

func (p *Permission) RemoveUser(userId uint64) error {
	return Delete(TablePermUsers, goqu.Ex{"perm_id": p.Id, "user_id": userId})
}

func GetGroupPermissions(groupId uint64) ([]Permission, error) {
	var perms []Permission
	sql, args, err := Goqu.From(TablePermissions).
		Where(goqu.I("group_id").Eq(groupId)).
		Order(goqu.I("priority").Desc(), goqu.I("id").Desc()).
		Prepared(true).
		ToSQL()
	if err != nil {
		return perms, err
	}
	err = Sqlx.Select(&perms, sql, args...)
	if err == nil {
		for i := range perms {
			perms[i].SetFullfilled()
			_ = perms[i].LoadEntities()
		}
	}
	return perms, err
}

func GetUserPermissionLevel(groupId, userId, channelId uint64, permName PermName) PermLevel {
	group, err := NewGroupFromId(groupId)
	if err == nil && !group.IsEmpty() && group.OwnerId == userId {
		return PermLevelAll
	}

	if permName != PermNameAdmin {
		adminLevel := GetUserPermissionLevel(groupId, userId, 0, PermNameAdmin)
		if adminLevel == PermLevelAll {
			return PermLevelAll
		}
	}

	var roleIds []uint64
	member, _ := NewMemberFromId(groupId, userId)
	if !member.IsEmpty() {
		for _, r := range member.Roles() {
			roleIds = append(roleIds, r.Id)
		}
	}
	everyoneRole, errEveryone := NewRoleFromName("@everyone", groupId)
	if errEveryone == nil && !everyoneRole.IsEmpty() {
		alreadyHas := false
		for _, rid := range roleIds {
			if rid == everyoneRole.Id {
				alreadyHas = true
				break
			}
		}
		if !alreadyHas {
			roleIds = append(roleIds, everyoneRole.Id)
		}
	}

	orConds := []goqu.Expression{
		goqu.I("app_perms.id").In(
			Goqu.From(TablePermUsers).Select("perm_id").Where(goqu.I("user_id").Eq(userId)),
		),
	}
	if len(roleIds) > 0 {
		orConds = append(orConds, goqu.I("app_perms.id").In(
			Goqu.From(TablePermRoles).Select("perm_id").Where(goqu.I("role_id").In(roleIds)),
		))
	}
	if channelId != 0 {
		orConds = append(orConds, goqu.I("app_perms.id").In(
			Goqu.From(TablePermChannels).Select("perm_id").Where(goqu.I("channel_id").Eq(channelId)),
		))
	}

	var matchingPerm Permission
	sqlStr, args, err := Goqu.From(TablePermissions).
		Where(
			goqu.I("app_perms.group_id").Eq(groupId),
			goqu.I("app_perms.name").Eq(permName),
			goqu.Or(orConds...),
		).
		Order(
			goqu.I("app_perms.priority").Desc(),
			goqu.I("app_perms.id").Desc(),
		).
		Limit(1).
		Prepared(true).
		ToSQL()

	if err == nil {
		if errGet := Sqlx.Get(&matchingPerm, sqlStr, args...); errGet == nil {
			matchingPerm.SetFullfilled()
			return matchingPerm.Level
		}
	}

	switch permName {
	case PermNameRead, PermNameWrite, PermNameDelete, PermNameMembers:
		return PermLevelOwn
	case PermNameRoles, PermNameSettings, PermNameAdmin:
		return PermLevelDisable
	default:
		return PermLevelDisable
	}
}
