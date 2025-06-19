package model_database

import "github.com/Mopsgamer/draqun/server/environment"

const (
	RegexpMessageContent = `^.+$`
)

func IsValidMessageContent(text string) error {
	if IsValidString(text, RegexpMessageContent, 8000) {
		return environment.ErrChatMessageContent
	}

	return nil
}
