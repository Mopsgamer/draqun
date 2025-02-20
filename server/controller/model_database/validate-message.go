package model_database

const (
	RegexpMessageContent = `^.+$`
)

func IsValidMessageContent(text string) bool {
	return IsValidString(text, RegexpMessageContent, 8000)
}
