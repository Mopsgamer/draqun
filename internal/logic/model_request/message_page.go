package model_request

type MessagesPage struct {
	*GroupIdUri
	Page uint64 `uri:"messages_page"`
}
