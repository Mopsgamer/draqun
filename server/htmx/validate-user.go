package htmx

import (
	"regexp"
)

var (
	RegexpUserPassword = regexp.MustCompile(`^.{8,255}$`)
	RegexpUserNick     = regexp.MustCompile(`^.{1,255}$`)
	RegexpUserName     = regexp.MustCompile(`^[a-zA-Z0-9._]{1,255}$`)
	// Source: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/email#basic_validation
	RegexpUserEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	RegexpUserPhone = regexp.MustCompile(`^(|\+?\d{6,14})$`)
)

func IsValidUserPassword(str string) error {
	if !RegexpUserPassword.Match([]byte(str)) {
		return ErrFormatUserPassword
	}
	return nil
}

func IsValidUserNick(str string) error {
	if !RegexpUserNick.Match([]byte(str)) {
		return ErrFormatUserMoniker
	}
	return nil
}

func IsValidUserName(str string) error {
	if !RegexpUserName.Match([]byte(str)) {
		return ErrFormatUserName
	}
	return nil
}

func IsValidUserEmail(str string) error {
	if !RegexpUserEmail.Match([]byte(str)) {
		return ErrFormatUserEmail
	}
	return nil
}

func IsValidUserPhone(str string) error {
	str = regexp.MustCompile(`[\s-]`).ReplaceAllString(str, "")
	if !RegexpUserPhone.Match([]byte(str)) {
		return ErrFormatUserPhone
	}
	return nil
}
