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
	return First[model_database.Message](db, "app_group_messages", goqu.Ex{"id": messageId})
}

func (db Database) MessageCreate(message model_database.Message) *uint64 {
	return Insert(db, "app_group_messages", message)
}

func (db Database) MessageList(groupId uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ?`
	err := db.Sqlx.Select(messageList, query, groupId)

	if err != nil {
		if err == sql.ErrNoRows {
			return *messageList
		}
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db *Database) MessageFirst(groupId uint64) *model_database.Message {
	message := new(model_database.Message)
	query := `SELECT * FROM app_group_messages
		WHERE group_id = ?
		ORDER BY id ASC LIMIT 1`
	err := db.Sqlx.Get(message, query, groupId)

	if err != nil {
		log.Error(err)
		return message
	}
	return message
}

func (db *Database) MessageLast(groupId uint64) *model_database.Message {
	message := new(model_database.Message)
	query := `SELECT * FROM app_group_messages
		WHERE group_id = ?
		ORDER BY id DESC LIMIT 1`
	err := db.Sqlx.Get(message, query, groupId)

	if err != nil {
		log.Error(err)
		return message
	}
	return message
}

func (db Database) MessageListPage(groupId uint64, page uint64, perPage uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM (
		SELECT * FROM app_group_messages
		WHERE group_id = ?
		ORDER BY id DESC
		LIMIT ?, ?
	) subquery
	ORDER BY id ASC`
	from := (page - 1) * perPage
	to := page * perPage
	err := db.Sqlx.Select(messageList, query, groupId, from, to)

	if err != nil {
		if err == sql.ErrNoRows {
			return *messageList
		}
		log.Error(err)
		return *messageList
	}
	return *messageList
}
