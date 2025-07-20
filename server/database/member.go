package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) MemberById(groupId, userId uint64) *model_database.Member {
	return First[model_database.Member](db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberCreate(member model_database.Member) bool {
	return Insert(db, TableMembers, member) != nil
}

func (db Database) MemberDelete(userId, groupId uint64) bool {
	return Delete(db, TableMembers, exp.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberList(groupId uint64) []model_database.User {
	memberList := new([]model_database.User)

	err := db.Goqu.From(TableMembers).Select(TableUsers+".*").Prepared(true).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		ScanStructs(memberList)

	if err == sql.ErrNoRows {
		return *memberList
	}

	if err != nil {
		log.Error(err)
	}

	return *memberList
}

func (db Database) MemberListPage(groupId uint64, page, perPage uint) []model_database.User {
	memberList := new([]model_database.User)
	from := (page - 1) * perPage

	err := db.Goqu.Select(TableUsers+".*").From(TableMembers).
		LeftJoin(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableMembers+".user_id")))).
		Where(goqu.Ex{TableMembers + ".group_id": groupId}).
		Order(goqu.I(TableMembers + ".user_id").Asc()).
		Limit(perPage).Offset(from).
		ScanStructs(memberList)

	if err == sql.ErrNoRows {
		return *memberList
	}

	if err != nil {
		log.Error(err)
	}

	return *memberList
}
