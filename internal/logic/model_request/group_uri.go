package model_request

type GroupUri struct {
	GroupId   uint64 `uri:"group_id"`
	GroupName string `uri:"group_name"`
}
