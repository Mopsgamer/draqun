package model_request

type GroupIdUri struct {
	GroupId uint64 `uri:"group_id"`
}

type GroupNameUri struct {
	GroupName string `uri:"group_name"`
}
