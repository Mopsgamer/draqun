package htmx

const (
	RegexpMessageContent = `^.+$`
)

func IsValidMessageContent(text string) error {
	if IsValidString(text, RegexpMessageContent, 8000) {
		return ErrGroupChatInvalidContent
	}

	return nil
}
