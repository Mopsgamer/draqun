package model

import "regexp"

const (
	RegexpPassword        string = "^[a-zA-Z0-9, .~\\-+%$^&*_!?()[\\]{}`]{8,255}$"
	MessageDetailPassword string = "Must contain only letters (A-Z, a-z), numbers (0-9), spaces, or these special characters: , . ~ - + % $ ^ & * _ ! ? ( ) [ ] { } `. Must be at least 8 characters long and no more than 255 characters."

	RegexpNickname        string = "^.{1,255}$"
	MessageDetailNickname string = "Must be between 1 and 255 characters long and can contain any characters."

	RegexpUsername        string = "^[a-zA-Z0-9._]{1,255}$"
	MessageDetailUsername string = "Must contain only letters (A-Z, a-z), numbers (0-9), and these special characters: . _ . No spaces. Must be at least 1 characters long and no more than 255 characters."

	// Source: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/email#basic_validation
	RegexpEmail        string = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	MessageDetailEmail string = "Must be a valid email."

	RegexpPhone        string = "^()$"
	MessageDetailPhone string = "Must be a valid phone number."
)

func ValidateString(str string, rg string) bool {
	l := len(str)
	return l >= 0 && l <= 255 && regexp.MustCompile(rg).MatchString(str)
}

func ValidatePassword(str string) bool {
	return ValidateString(str, RegexpPassword)
}
func ValidateNickname(str string) bool {
	return ValidateString(str, RegexpNickname)
}
func ValidateUsername(str string) bool {
	return ValidateString(str, RegexpUsername)
}
func ValidateEmail(str string) bool {
	return ValidateString(str, RegexpEmail)
}
