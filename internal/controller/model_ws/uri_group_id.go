package model_ws

import (
	"restapp/internal/controller/controller_ws"
	"restapp/internal/controller/model_database"
)

type UriGroupId struct {
	GroupId uint64 `uri:"group_id"`
}

func (request *UriGroupId) Group(ctl *controller_ws.ControllerWs) *model_database.Group {
	return ctl.Group
}
