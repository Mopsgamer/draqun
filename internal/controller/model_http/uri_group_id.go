package model_http

import (
	"restapp/internal/controller/controller_http"
	"restapp/internal/controller/model_database"
)

type UriGroupId struct {
	GroupId uint64 `uri:"group_id"`
}

func (request *UriGroupId) Group(ctl controller_http.ControllerHttp) *model_database.Group {
	return ctl.DB.GroupById(request.GroupId)
}
