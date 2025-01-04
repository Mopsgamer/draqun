package model_http

type MembersPage struct {
	*GroupIdUri
	Page uint64 `uri:"members_page"`
}
