package model_http

import (
	"github.com/Mopsgamer/vibely/internal/controller/controller_http"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
)

type UriGroupName struct {
	GroupName string `uri:"group_name"`
}

func (request *UriGroupName) Group(ctl controller_http.ControllerHttp) *model_database.Group {
	return ctl.DB.GroupByName(request.GroupName)
}
