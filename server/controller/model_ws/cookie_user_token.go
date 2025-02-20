package model_ws

import (
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/controller/model_database"
)

type CookieUserToken struct {
	UserToken string `cookie:"Authorization"`
}

// Get owner of the request using the "Authorization" header.
func (request *CookieUserToken) User(ctl *controller_ws.ControllerWs) *model_database.User {
	return ctl.User
}
