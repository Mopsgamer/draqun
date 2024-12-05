package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) MessageById(messageId uint64) *model_database.Message {
	message := new(model_database.Message)
	query := `SELECT * FROM app_group_messages WHERE id = ?`
	err := db.Sql.Get(message, query, messageId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return message
}

func (db Database) MessageCreate(message model_database.Message) *uint64 {
	query :=
		`INSERT INTO app_group_messages (
			group_id,
			author_id,
			content,
			created_at
		)
    	VALUES (?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		message.GroupId,
		message.AuthorId,
		message.Content,
		message.CreatedAt,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}

func (db Database) MessageList(groupId uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ?`
	err := db.Sql.Select(messageList, query, groupId)

	if err != nil {
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
	err := db.Sql.Get(message, query, groupId)

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
	err := db.Sql.Get(message, query, groupId)

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
	err := db.Sql.Select(messageList, query, groupId, from, to)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}
