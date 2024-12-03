package model_request

type MessagesPage struct {
	*GroupUri
	Page uint64 `uri:"messages_page"`
}
