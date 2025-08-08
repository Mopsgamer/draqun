package model

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type Member struct {
	Db *DB `db:"-"`

	GroupId           uint64        `db:"group_id"`
	UserId            uint64        `db:"user_id"`
	Moniker           string        `db:"moniker"`
	FirstTimeJoinedAt time.Time     `db:"first_time_joined_at"`
	IsDeleted         types.BitBool `db:"is_deleted"`
}

func NewMember(db *DB, groupId, userId uint64, moniker string) Member {
	return Member{
		Db:                db,
		GroupId:           groupId,
		UserId:            userId,
		Moniker:           moniker,
		FirstTimeJoinedAt: time.Now(),
	}
}

func NewMemberFromId(db *DB, groupId, userId uint64) (bool, Member) {
	member := Member{Db: db}
	return member.FromId(groupId, userId), member
}

func (member Member) IsEmpty() bool {
	return member.GroupId != 0 && member.UserId != 0
}

func (member Member) IsAvailable() bool {
	return !member.IsEmpty() && !bool(member.IsDeleted)
}

func (member *Member) Insert() error {
	return Insert0(member.Db, TableMembers, member)
}

func (member Member) Update() error {
	return Update(member.Db, TableMembers, member, goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId})
}

func (member *Member) FromId(groupId, userId uint64) bool {
	First(member.Db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId}, member)
	return member.IsEmpty()
}

func (member *Member) User() User {
	user := User{Db: member.Db}
	user.FromId(member.UserId)
	return user
}

func (member Member) Group() Group {
	group := Group{Db: member.Db}
	group.FromId(member.GroupId)
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

func (member Member) Ban(creatorId uint64, endsAt time.Time, description string) error {
	action := ActionBan{
		Db:          member.Db,
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     time.Now(),
		EndsAt:      endsAt,
	}

	return action.Insert()
}

func (member Member) Unban(revokerId uint64) error {
	ban := ActionBan{Db: member.Db}
	ban.FromId(member.UserId, member.GroupId)
	if ban.IsEmpty() {
		return nil
	}

	ban.RevokerId = revokerId
	return ban.Update()
}

func (member Member) Kick(creatorId uint64, description string) error {
	action := ActionKick{
		Db:          member.Db,
		GroupId:     member.GroupId,
		TargetId:    member.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     time.Now(),
	}

	return action.Insert()
}

func (member Member) LeaveActed() error {
	action := ActionMembership{
		Db:      member.Db,
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  false,
		ActedAt: time.Now(),
	}

	return action.Insert()
}

func (member Member) JoinActed() error {
	action := ActionMembership{
		Db:      member.Db,
		GroupId: member.GroupId,
		UserId:  member.UserId,
		IsJoin:  false,
		ActedAt: time.Now(),
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
