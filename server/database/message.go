package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) CachedMessageList(messageList []model_database.Message) []fiber.Map {
	result := []fiber.Map{}
	users := map[uint64]model_database.User{}

	for _, message := range messageList {
		if _, ok := users[message.AuthorId]; !ok {
			users[message.AuthorId] = *db.UserById(message.AuthorId)
		}
		author := users[message.AuthorId]
		result = append(result, fiber.Map{
			"Message": message,
			"Author":  author,
		})
	}

	return result
}

func (db Database) MessageById(messageId uint64) *model_database.Message {
	return First[model_database.Message](db, TableMessages, goqu.Ex{"id": messageId})
}

func (db Database) MessageCreate(message model_database.Message) *uint64 {
	return Insert(db, TableMessages, message)
}

func (db Database) MessageList(groupId uint64) []model_database.Message {
	messageList := new([]model_database.Message)

	err := db.Goqu.From(TableMessages).Where(goqu.C("group_id").Eq(groupId)).
		ScanStructs(messageList)

	if err == sql.ErrNoRows {
		return *messageList
	}

	if err != nil {
		log.Error(err)
	}

	return *messageList
}

func (db *Database) MessageFirst(groupId uint64) *model_database.Message {
	message := new(model_database.Message)

	found, err := db.Goqu.From(TableMessages).Prepared(true).Where(goqu.C("group_id").Eq(groupId)).Order(goqu.C("id").Asc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (db *Database) MessageLast(groupId uint64) *model_database.Message {
	message := new(model_database.Message)

	found, err := db.Goqu.From(TableMessages).Prepared(true).Where(goqu.C("group_id").Eq(groupId)).Order(goqu.C("id").Desc()).Limit(1).
		ScanStruct(message)

	if !found {
		log.Error(err)
	}

	return message
}

func (db Database) MessageListPage(groupId uint64, page, perPage uint) []model_database.Message {
	messageList := new([]model_database.Message)
	from := (page - 1) * perPage

	subquery := db.Goqu.From(TableMessages).
		Where(goqu.Ex{"group_id": groupId}).
		Order(goqu.I("id").Desc()).
		Limit(perPage).Offset(from)
	err := db.Goqu.From(subquery.As("subquery")).Order(goqu.I("id").Asc()).
		Executor().ScanStructs(messageList)

	if err == sql.ErrNoRows {
		return *messageList
	}

	if err != nil {
		log.Error(err)
	}

	return *messageList
}
