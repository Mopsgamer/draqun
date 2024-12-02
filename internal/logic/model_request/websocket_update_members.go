package model_request

type WebsocketUpdateMembers struct {
	*WebsocketRequest
	MemberId uint64 `json:"member-id"`
}
