package logic_websocket

import (
	"restapp/internal/logic/database"
	"restapp/internal/logic/model_database"
)

func (ws LogicWebsocket) GroupJoin(member model_database.Member) error {
	return GroupJoin(*ws.DB, member)
}

func (ws LogicWebsocket) MessageSend(message model_database.Message) *uint64 {
	return MessageSend(*ws.DB, message)
}

func GroupJoin(db database.Database, member model_database.Member) error {
	err := db.MemberCreate(member)
	return err
}

func MessageSend(db database.Database, message model_database.Message) *uint64 {
	messageId := db.MessageCreate(message)
	return messageId
}
