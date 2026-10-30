package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var MentionEntityRegex = regexp.MustCompile(`(?i)@(u|r|c|g|user|role|channel|group):(\d+)`)

// ExtractMentionedUsers finds all users in group mentioned in content by ID-based entity syntax
func ExtractMentionedUsers(groupId uint64, content string) ([]User, error) {
	return ExtractMentionedUsersForMessage(groupId, 0, content)
}

// ExtractMentionedUsersForMessage finds all users mentioned in content, checking group membership and channel read permissions
func ExtractMentionedUsersForMessage(groupId, channelId uint64, content string) ([]User, error) {
	matches := MentionEntityRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	group, err := NewGroupFromId(groupId)
	if err != nil || group.IsEmpty() {
		return nil, err
	}

	userMap := make(map[uint64]User)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		entityType := strings.ToLower(match[1])
		idVal, err := strconv.ParseUint(match[2], 10, 64)
		if err != nil || idVal == 0 {
			continue
		}

		switch entityType {
		case "u", "user":
			mb, errMb := NewMemberFromId(groupId, idVal)
			if errMb == nil && !mb.IsEmpty() && !mb.IsDeleted {
				u, errU := NewUserFromId(idVal)
				if errU == nil && !u.IsEmpty() && !u.IsSuspended {
					if channelId == 0 || GetUserPermissionLevel(groupId, u.Id, channelId, PermNameRead) != PermLevelDisable {
						userMap[u.Id] = u
					}
				}
			}

		case "r", "role":
			role, errR := NewRoleFromId(idVal, groupId)
			if errR == nil && !role.IsEmpty() {
				var roleUsers []User
				if role.Name == "@everyone" {
					roleUsers = group.UsersPage(1, 10000)
				} else {
					roleUsers = role.Assignees()
				}
				for _, u := range roleUsers {
					if u.IsSuspended {
						continue
					}
					mb, errMb := NewMemberFromId(groupId, u.Id)
					if errMb != nil || mb.IsEmpty() || mb.IsDeleted {
						continue
					}
					if channelId == 0 || GetUserPermissionLevel(groupId, u.Id, channelId, PermNameRead) != PermLevelDisable {
						userMap[u.Id] = u
					}
				}
			}

		case "g", "group":
			if idVal == groupId {
				allMembers := group.UsersPage(1, 10000)
				for _, u := range allMembers {
					if u.IsSuspended {
						continue
					}
					if channelId == 0 || GetUserPermissionLevel(groupId, u.Id, channelId, PermNameRead) != PermLevelDisable {
						userMap[u.Id] = u
					}
				}
			}
		}
	}

	var result []User
	for _, u := range userMap {
		result = append(result, u)
	}

	return result, nil
}

// ProcessMessageMentions parses message for mentions, inserts notifications, and broadcasts WebSocket push
func ProcessMessageMentions(app *fiber.App, msg Message) {
	mentionedUsers, err := ExtractMentionedUsersForMessage(msg.GroupId, msg.ChannelId, string(msg.Content))
	if err != nil || len(mentionedUsers) == 0 {
		return
	}

	author := msg.Author()
	group, _ := NewGroupFromId(msg.GroupId)
	channel, _ := NewChannelFromId(msg.ChannelId)

	for _, targetUser := range mentionedUsers {
		if targetUser.Id == msg.AuthorId {
			continue
		}

		title := fmt.Sprintf("%s mentioned you in %s", author.Moniker, group.Moniker)
		if !channel.IsEmpty() {
			title = fmt.Sprintf("%s mentioned you in #%s (%s)", author.Moniker, channel.Moniker, group.Moniker)
		}

		notif := NewNotification(
			targetUser.Id,
			msg.AuthorId,
			msg.GroupId,
			msg.ChannelId,
			msg.Id,
			"mention",
			title,
			string(msg.Content),
		)

		_ = CreateAndPushNotification(app, notif)
	}
}
