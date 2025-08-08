package htmx

import "regexp"

var (
	RegexpMessageContent = regexp.MustCompile(`^.+$`)
)

func IsValidMessageContent(str string) error {
	if !RegexpMessageContent.Match([]byte(str)) {
		return ErrGroupChatInvalidContent
	}

	return nil
}
