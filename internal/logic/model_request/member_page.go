package model_request

type MembersPage struct {
	*GroupUri
	Page uint64 `uri:"members_page"`
}
