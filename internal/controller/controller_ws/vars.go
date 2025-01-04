package controller_ws

import "restapp/internal/controller/model_database"

// Get owner of the request using the initial websocket connection URI.
func (ws *ControllerWs) User() (user *model_database.User) {
	if user, ok := (*ws.Map)["User"].(*model_database.User); ok {
		return user
	}

	ws.Closed = true
	return nil
}

// Get group by the id from initial websocket connection URI.
func (ws *ControllerWs) Group() *model_database.Group {
	if group, ok := (*ws.Map)["Group"].(*model_database.Group); ok {
		return group
	}
	return nil
}

func (ws *ControllerWs) Member() *model_database.Member {
	if member, ok := (*ws.Map)["Member"].(*model_database.Member); ok {
		return member
	}
	return nil
}

func (ws *ControllerWs) Rights() *model_database.Role {
	if rights, ok := (*ws.Map)["Rights"].(*model_database.Role); ok {
		return rights
	}
	return nil
}
