package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type Member struct {
	checkEmpty
	GroupId           uint64   `db:"group_id"`
	UserId            uint64   `db:"user_id"`
	Moniker           Moniker  `db:"moniker"`
	FirstTimeJoinedAt TimePast `db:"first_time_joined_at"`
	IsDeleted         bool     `db:"is_deleted"`
}

var _ Model = (*Member)(nil)

func NewMember(groupId, userId uint64, moniker Moniker) Member {
	return Member{
		GroupId:           groupId,
		UserId:            userId,
		Moniker:           moniker,
		FirstTimeJoinedAt: TimePast(time.Now()),
	}
}

func NewMemberFromId(groupId, userId uint64) (Member, error) {
	member := Member{}
	return member, First(TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId}, &member)
}

func (member Member) Validate() htmx.Alert {
	if !member.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !member.FirstTimeJoinedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (member Member) IsAvailable() bool {
	return !member.IsEmpty() && !bool(member.IsDeleted)
}

func (member *Member) Insert() error {
	return Insert(TableMembers, member)
}

func (member Member) Update() error {
	return Update(TableMembers, member, goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId})
}

func (member *Member) User() User {
	user, _ := NewUserFromId(member.UserId)
	return user
}

func (member Member) Group() Group {
	group, _ := NewGroupFromId(member.GroupId)
	return group
}

func (member Member) Roles() []Role {
	roleList := []Role{}
	sql, args, err := Goqu.From(TableRoles).Select(TableRoles+".*").
		LeftJoin(goqu.T(TableRoleAssignees), goqu.On(
			goqu.I(TableRoleAssignees+".role_id").Eq(goqu.C("id").Table(TableRoles)),
		)).
		Where(goqu.Ex{TableRoles + ".group_id": member.GroupId, TableRoleAssignees + ".user_id": member.UserId}).
		Order(goqu.I(TableRoles+".position").Asc(), goqu.I(TableRoles+".id").Asc()).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return roleList
	}

	err = Sqlx.Select(&roleList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return roleList
}

func (member Member) Role() Role {
	group := member.Group()
	roleList := member.Roles()
	if len(roleList) > 0 {
		return roleList[0]
	}
	return group.Everyone()
}

func (member Member) Ban(creatorId uint64, endsAt TimeFuture, description Description) error {
	action := ActionBan{
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     TimePast(time.Now()),
		EndsAt:      endsAt,
	}

	if err := action.Insert(); err != nil {
		return err
	}

	member.IsDeleted = true
	if err := member.Update(); err != nil {
		return err
	}

	roleIds := []uint64{}
	for _, r := range member.Group().Roles() {
		roleIds = append(roleIds, r.Id)
	}
	if len(roleIds) > 0 {
		_ = Delete(TableRoleAssignees, goqu.Ex{"user_id": member.UserId, "role_id": roleIds})
	}

	return nil
}

func (member Member) Unban(revokerId uint64) error {
	ban, err := NewActionBanFromId(member.UserId, member.GroupId)
	if err != nil {
		return err
	}

	ban.RevokerId = revokerId
	return ban.Update()
}

func (member Member) Kick(creatorId uint64, description Description) error {
	action := ActionKick{
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     TimePast(time.Now()),
	}

	if err := action.Insert(); err != nil {
		return err
	}

	member.IsDeleted = true
	if err := member.Update(); err != nil {
		return err
	}

	roleIds := []uint64{}
	for _, r := range member.Group().Roles() {
		roleIds = append(roleIds, r.Id)
	}
	if len(roleIds) > 0 {
		_ = Delete(TableRoleAssignees, goqu.Ex{"user_id": member.UserId, "role_id": roleIds})
	}

	return nil
}

func (member Member) LeaveActed() error {
	action := ActionMembership{
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  false,
		ActedAt: TimePast(time.Now()),
	}

	return action.Insert()
}

func (member Member) JoinActed() error {
	action := ActionMembership{
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  true,
		ActedAt: TimePast(time.Now()),
	}

	return action.Insert()
}

func (member Member) LastTimeJoinedAt() TimePast {
	action := ActionMembership{}
	sql, args, err := Goqu.From(TableMemberships).
		Where(goqu.Ex{
			"group_id": member.GroupId,
			"user_id":  member.UserId,
			"is_join":  true,
		}).
		Order(goqu.I("acted_at").Desc()).
		Limit(1).
		Prepared(true).ToSQL()
	if err != nil {
		return member.FirstTimeJoinedAt
	}

	err = Sqlx.Get(&action, sql, args...)
	if err != nil || time.Time(action.ActedAt).IsZero() {
		return member.FirstTimeJoinedAt
	}

	return action.ActedAt
}

func (member Member) MessagesCount() uint64 {
	count := uint64(0)
	sql, args, err := Goqu.Select(goqu.COUNT("*").As("count")).From(TableMessages).
		Where(goqu.Ex{"group_id": member.GroupId, "author_id": member.UserId}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return count
	}

	err = Sqlx.Get(&count, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return count
}

func (member Member) ActionListPage(page uint, limit uint) []Action {
	actions := []Action{}
	from := (page - 1) * limit
	filter := goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId}
	sql, args, err := Goqu.From(TableBans).UnionAll(
		Goqu.From(TableKicks).UnionAll(
			Goqu.From(TableMemberships).Where(filter),
		).Where(filter),
	).Where(filter).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return actions
	}

	err = Sqlx.Select(&actions, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return actions
}
