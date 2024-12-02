package model_request

type WebsocketUpdateChat struct {
	*WebsocketRequest
	MessageId uint64 `json:"message-id,string"`
}
