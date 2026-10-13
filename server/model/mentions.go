package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9._]+)`)

// ExtractMentionedUsers finds all users in group mentioned in content by @name or @moniker
func ExtractMentionedUsers(groupId uint64, content string) ([]User, error) {
	matches := mentionRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	group, err := NewGroupFromId(groupId)
	if err != nil || group.IsEmpty() {
		return nil, err
	}

	members := group.UsersPage(1, 1000)
	userMap := make(map[uint64]User)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		target := strings.ToLower(match[1])

		for _, u := range members {
			uName := strings.ToLower(string(u.Name))
			uMoniker := strings.ToLower(string(u.Moniker))

			if uName == target || uMoniker == target {
				userMap[u.Id] = u
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
	mentionedUsers, err := ExtractMentionedUsers(msg.GroupId, string(msg.Content))
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
