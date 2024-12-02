package logic_websocket

import "restapp/internal/logic/model_database"

func UserIsOnline(user *model_database.User) bool {
	if user == nil {
		return false
	}

	cons := *WebsocketConnections.Users
	arr, ok := cons[user.Id]
	return ok && len(arr) > 0
}
