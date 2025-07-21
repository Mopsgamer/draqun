package database

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

type Member struct {
	Db *goqu.Database

	GroupId           uint64    `db:"group_id"`
	UserId            uint64    `db:"user_id"`
	Moniker           string    `db:"moniker"`
	FirstTimeJoinedAt time.Time `db:"first_time_joined_at"`
}

func NewMemberEmpty(db *goqu.Database, groupId, userId uint64) Member {
	return Member{Db: db, GroupId: groupId, UserId: userId, FirstTimeJoinedAt: time.Now()}
}

func (member *Member) IsEmpty() bool {
	return member.GroupId != 0 && member.UserId != 0
}

func (member *Member) Insert() bool {
	id := Insert(member.Db, TableMembers, member)
	return id != nil
}

func (member *Member) Update() bool {
	return Update(member.Db, TableMembers, member, exp.Ex{"group_id": member.GroupId, "user_id": member.UserId})
}

func (member *Member) Delete(userId, groupId uint64) bool {
	return Delete(member.Db, TableMembers, exp.Ex{"group_id": groupId, "user_id": userId})
}

func (member *Member) FromId(groupId, userId uint64) bool {
	First(member.Db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId}, member)
	return member.IsEmpty()
}

func (member *Member) User() User {
	user := NewUser(member.Db)
	user.FromId(member.UserId)
	return user
}

func (member *Member) Group() Group {
	group := Group{Db: member.Db}
	group.FromId(member.GroupId)
	return group
}

func (member *Member) Roles() []Role {
	roleList := new([]Role)
	err := member.Db.From(TableRoles).Select(TableRoles+".*").
		LeftJoin(goqu.T(TableRoleAssigns), goqu.On(goqu.I(TableRoleAssigns+".role_id").Eq(goqu.I(TableRoles+".id")))).
		Where(goqu.Ex{TableRoleAssigns + ".group_id": member.GroupId, TableRoleAssigns + ".user_id": member.UserId}).
		ScanStructs(roleList)

	if err == sql.ErrNoRows {
		return *roleList
	}

	if err != nil {
		log.Error(err)
	}

	return *roleList
}

func (member *Member) Role() Role {
	roleList := member.Roles()
	group := member.Group()
	everyone := group.Everyone()
	if len(roleList) == 0 {
		return everyone
	}

	everyone.Merge(roleList...)
	return everyone
}
