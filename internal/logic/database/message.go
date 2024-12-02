package database

import (
	"math"
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

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

func (db *Database) MessageLast(groupId uint64) *model_database.Message {
	messageList := new(model_database.Message)
	query := `SELECT * FROM app_group_messages
		WHERE group_id = ?
		ORDER BY id DESC LIMIT 1`
	err := db.Sql.Select(messageList, query, groupId)

	if err != nil {
		log.Error(err)
		return messageList
	}
	return messageList
}

func (db *Database) MessageListAround(groupId uint64, messageId uint64, radius uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ? AND message_id > ? AND message_id < ?`
	radiusMin := max(0, radius)
	radiusMax := min(math.MaxUint64, radius)
	err := db.Sql.Select(messageList, query, groupId, messageId-radiusMin, messageId+radiusMax)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db Database) MessageListAfter(groupId uint64, afterMessageId uint64, count uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ? AND message_id > ? LIMIT ?`
	err := db.Sql.Select(messageList, query, groupId, afterMessageId, count)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db Database) MessageListBefore(groupId uint64, afterMessageId uint64, count uint64) []model_database.Message {
	messageList := &[]model_database.Message{}
	query := `SELECT * FROM (
			SELECT * FROM app_group_messages
			WHERE group_id = ? AND id < ?
			ORDER BY id DESC LIMIT ?
		) subquery
        ORDER BY id ASC`
	err := db.Sql.Select(messageList, query, groupId, afterMessageId, count)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}
