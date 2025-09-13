package model

import (
	"time"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type Member struct {
	Db *DB `db:"-"`

	GroupId           uint64        `db:"group_id"`
	UserId            uint64        `db:"user_id"`
	Moniker           Moniker       `db:"moniker"`
	FirstTimeJoinedAt TimePast      `db:"first_time_joined_at"`
	IsDeleted         types.BitBool `db:"is_deleted"`
}

var _ Model = (*Member)(nil)

func NewMember(db *DB, groupId, userId uint64, moniker Moniker) Member {
	return Member{
		Db:                db,
		GroupId:           groupId,
		UserId:            userId,
		Moniker:           moniker,
		FirstTimeJoinedAt: TimePast(time.Now()),
	}
}

func NewMemberFromId(db *DB, groupId, userId uint64) (Member, error) {
	member := Member{Db: db}
	err := member.FromId(groupId, userId)
	return member, err
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

func (member Member) IsEmpty() bool {
	return member.GroupId == 0 || member.UserId == 0
}

func (member Member) IsAvailable() bool {
	return !member.IsEmpty() && !bool(member.IsDeleted)
}

func (member *Member) Insert() error {
	return Insert(member.Db, TableMembers, member)
}

func (member Member) Update() error {
	return Update(member.Db, TableMembers, member, goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId})
}

func (member *Member) FromId(groupId, userId uint64) error {
	return First(member.Db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId}, member)
}

func (member *Member) User() User {
	user := User{Db: member.Db}
	_ = user.FromId(member.UserId)
	return user
}

func (member Member) Group() Group {
	group := Group{Db: member.Db}
	_ = group.FromId(member.GroupId)
	return group
}

func (member Member) Roles() []Role {
	roleList := []Role{}
	sql, args, err := member.Db.Goqu.From(TableRoles).Select(TableRoles+".*").
		LeftJoin(goqu.T(TableRoleAssignees), goqu.On(goqu.I(TableRoleAssignees+".role_id").Eq(TableRoles+".id"))).
		Where(goqu.Ex{TableRoles + ".group_id": member.GroupId, TableRoleAssignees + ".user_id": member.UserId}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return roleList
	}

	err = member.Db.Sqlx.Select(&roleList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return roleList
}

func (member Member) Role() Role {
	roleList := member.Roles()
	group := member.Group()
	everyone := group.Everyone()
	if len(roleList) == 0 {
		return everyone
	}

	everyone.Merge(roleList...)
	return everyone
}

func (member Member) Ban(creatorId uint64, endsAt TimeFuture, description Description) error {
	action := ActionBan{
		Db:          member.Db,
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     TimePast(time.Now()),
		EndsAt:      endsAt,
	}

	return action.Insert()
}

func (member Member) Unban(revokerId uint64) error {
	ban := ActionBan{Db: member.Db}
	err := ban.FromId(member.UserId, member.GroupId)
	if err != nil {
		return err
	}

	ban.RevokerId = revokerId
	return ban.Update()
}

func (member Member) Kick(creatorId uint64, description Description) error {
	action := ActionKick{
		Db:          member.Db,
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     TimePast(time.Now()),
	}

	return action.Insert()
}

func (member Member) LeaveActed() error {
	action := ActionMembership{
		Db:      member.Db,
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  false,
		ActedAt: TimePast(time.Now()),
	}

	return action.Insert()
}

func (member Member) JoinActed() error {
	action := ActionMembership{
		Db:      member.Db,
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  false,
		ActedAt: TimePast(time.Now()),
	}

	return action.Insert()
}

func (member Member) ActionListPage(page uint, limit uint) []Action {
	actions := []Action{}
	from := (page - 1) * limit
	filter := goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId}
	sql, args, err := member.Db.Goqu.From(TableBans).UnionAll(
		member.Db.Goqu.From(TableKicks).UnionAll(
			member.Db.Goqu.From(TableMemberships).Where(filter),
		).Where(filter),
	).Where(filter).
		Limit(limit).Offset(from).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return actions
	}

	err = member.Db.Sqlx.Select(&actions, sql, args...)
	if err != nil {
		handleErr(err)
	}

	return actions
}
