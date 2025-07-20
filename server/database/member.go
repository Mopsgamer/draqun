package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

type Member struct {
	Db *goqu.Database
	*model_database.Member
}

func NewMember(db *goqu.Database) Member {
	return Member{Db: db}
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
	member.Member = First[model_database.Member](member.Db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId})
	return member.IsEmpty()
}

func (member *Member) User() User {
	user := User{Db: member.Db}
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
	rights := Role{Db: member.Db, Role: &model_database.RoleDefault} // TODO: groups should use custom @everyone roles
	if len(roleList) == 0 {
		return rights
	}

	roles := make([]model_database.Role, len(roleList))
	for i, r := range roleList {
		roles[i] = *r.Role
	}
	rights.Merge(roles...)
	return rights
}
