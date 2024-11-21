package model

const (
	RegexpMessageContent = `^.{1,8000}$`
)

func IsValidMessageContent(text string) bool {
	return IsValidString(text, RegexpMessageContent, 8000)
}
