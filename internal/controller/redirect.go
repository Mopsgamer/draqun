package controller

import "fmt"

func PathRedirectGroup(groupId uint64) string {
	return "/chat/groups/" + fmt.Sprintf("%d", groupId)
}

func PathRedirectGroupJoin(groupName string) string {
	return "/chat/groups/join/" + groupName
}
