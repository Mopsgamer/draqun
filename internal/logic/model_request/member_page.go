package model_request

type MembersPage struct {
	*GroupIdUri
	Page uint64 `uri:"members_page"`
}
