package model_ws

import (
	"github.com/Mopsgamer/vibely/internal/controller/controller_ws"
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
)

type CookieUserToken struct {
	UserToken string `cookie:"Authorization"`
}

// Get owner of the request using the "Authorization" header.
func (request *CookieUserToken) User(ctl *controller_ws.ControllerWs) *model_database.User {
	return ctl.User
}
