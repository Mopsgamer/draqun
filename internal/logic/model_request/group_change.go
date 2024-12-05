package model_request

type GroupChange struct {
	*GroupUri
	Nick        string `form:"change-group-nickname"`
	Mode        string `form:"change-group-mode"`
	Description string `form:"change-group-description"`
	Avatar      string `form:"change-group-avatar"`
}
