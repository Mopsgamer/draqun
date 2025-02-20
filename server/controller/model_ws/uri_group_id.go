package model_ws

import (
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
)

type UriGroupId struct {
	GroupId uint64 `uri:"group_id"`
}

func (request *UriGroupId) Group(ctl *controller_ws.ControllerWs) *model_database.Group {
	return ctl.Group
}
