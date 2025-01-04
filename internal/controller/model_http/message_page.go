package model_http

type MessagesPage struct {
	*GroupIdUri
	Page uint64 `uri:"messages_page"`
}
