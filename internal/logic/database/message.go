package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) GroupMessageList(groupId uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ?`
	err := db.Sql.Select(messageList, query, groupId)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db Database) GroupMessageListAfter(groupId uint64, afterMessageId uint64, count int) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ? AND message_id > ? LIMIT ?`
	err := db.Sql.Select(messageList, query, groupId, afterMessageId, count)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db Database) GroupMessageListBefore(groupId uint64, afterMessageId uint64, count int) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM (
			SELECT * FROM app_group_messages
			WHERE group_id = ?
			AND id < ? ORDER BY id DESC LIMIT ?
		) subquery
        ORDER BY id ASC`
	err := db.Sql.Select(messageList, query, groupId, afterMessageId, count)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}
