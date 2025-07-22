package database

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx/types"
)

type Member struct {
	Db *goqu.Database `db:"-"`

	GroupId           uint64        `db:"group_id"`
	UserId            uint64        `db:"user_id"`
	Moniker           string        `db:"moniker"`
	FirstTimeJoinedAt time.Time     `db:"first_time_joined_at"`
	IsDeleted         types.BitBool `db:"is_deleted"`
}

func NewMember(db *goqu.Database, groupId, userId uint64, moniker string) Member {
	return Member{
		Db:                db,
		GroupId:           groupId,
		UserId:            userId,
		Moniker:           moniker,
		FirstTimeJoinedAt: time.Now(),
	}
}

func NewMemberFromId(db *goqu.Database, groupId, userId uint64) (bool, Member) {
	member := Member{Db: db}
	return member.FromId(groupId, userId), member
}

func (group Member) IsEmpty() bool {
	return group.GroupId != 0 && group.UserId != 0
}

func (group *Member) Insert() bool {
	return Insert(group.Db, TableMembers, group) != 0
}

func (group Member) Update() bool {
	return Update(group.Db, TableMembers, group, goqu.Ex{"group_id": group.GroupId, "user_id": group.UserId})
}

func (group *Member) FromId(groupId, userId uint64) bool {
	First(group.Db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId}, group)
	return group.IsEmpty()
}

func (group *Member) User() User {
	user := User{Db: group.Db}
	user.FromId(group.UserId)
	return user
}

func (member Member) Group() Group {
	group := Group{Db: member.Db}
	group.FromId(member.GroupId)
	return group
}

func (group Member) Roles() []Role {
	roleList := new([]Role)
	err := group.Db.From(TableRoles).Select(TableRoles+".*").
		LeftJoin(goqu.T(TableRoleAssignees), goqu.On(goqu.I(TableRoleAssignees+".role_id").Eq(TableRoles+".id"))).
		Where(goqu.Ex{TableRoles + ".group_id": group.GroupId, TableRoleAssignees + ".user_id": group.UserId}).
		ScanStructs(roleList)

	if err == sql.ErrNoRows {
		return *roleList
	}

	if err != nil {
		log.Error(err)
	}

	return *roleList
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

func (group Member) Ban(creatorId uint64, endsAt time.Time, description string) bool {
	action := ActionBan{
		Db:          group.Db,
		GroupId:     group.GroupId,
		TargetId:    group.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     time.Now(),
		EndsAt:      endsAt,
	}

	return action.Insert()
}

func (group Member) Unban(revokerId uint64) bool {
	ban := ActionBan{Db: group.Db}
	ban.FromId(group.UserId, group.GroupId)
	if ban.IsEmpty() {
		return false
	}

	ban.RevokerId = revokerId
	return ban.Update()
}

func (group Member) Kick(creatorId uint64, description string) bool {
	action := ActionKick{
		Db:          group.Db,
		GroupId:     group.GroupId,
		TargetId:    group.UserId,
		CreatorId:   creatorId,
		Description: description,
		ActedAt:     time.Now(),
	}

	return action.Insert()
}

func (group Member) LeaveActed() bool {
	action := ActionMembership{
		Db:      group.Db,
		GroupId: group.GroupId,
		UserId:  group.UserId,
		IsJoin:  false,
		ActedAt: time.Now(),
	}

	return action.Insert()
}

func (group Member) JoinActed() bool {
	action := ActionMembership{
		Db:      group.Db,
		GroupId: group.GroupId,
		UserId:  group.UserId,
		IsJoin:  false,
		ActedAt: time.Now(),
	}

	return action.Insert()
}

func (member Member) ActionListPage(page uint, limit uint) ([]Action, bool) {
	actions := new([]Action)
	from := (page - 1) * limit
	filter := goqu.Ex{"group_id": member.GroupId, "user_id": member.UserId}
	err := member.Db.From(TableBans).UnionAll(
		member.Db.From(TableKicks).UnionAll(
			member.Db.From(TableMemberships).Where(filter),
		).Where(filter),
	).Where(filter).
		Limit(limit).Offset(from).
		ScanStructs(actions)

	if err == sql.ErrNoRows {
		return *actions, true
	}

	if err != nil {
		log.Error(err)
		return nil, false
	}

	return *actions, true
}
