package htmx

const (
	RegexpMessageContent = `^.+$`
)

func IsValidMessageContent(text string) error {
	if IsValidString(text, RegexpMessageContent, 1, 8000) {
		return ErrGroupChatInvalidContent
	}

	return nil
}
