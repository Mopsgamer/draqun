package logic

import "fmt"

func PathRedirectGroup(groupId uint64) string {
	return "/chat/groups/" + fmt.Sprintf("%d", groupId)
}
