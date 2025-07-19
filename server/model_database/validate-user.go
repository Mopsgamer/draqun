package model_database

import (
	"regexp"

	"github.com/Mopsgamer/draqun/server/environment"
)

var ()

const (
	RegexpUserPassword string = "^[a-zA-Z0-9, .~\\-+%$^&*_!?()[\\]{}`]{8,255}$"
	RegexpUserNick     string = `^.{1,255}$`
	RegexpUserName     string = `^[a-zA-Z0-9._]{1,255}$`
	// Source: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input/email#basic_validation
	RegexpUserEmail string = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	RegexpUserPhone string = `^\+?[1-9]\d{1,14}$`
)

func IsValidUserPassword(str string) error {
	if !IsValidString(str, RegexpUserPassword, 255) {
		return environment.ErrFormatUserPassword
	}
	return nil
}

func IsValidUserNick(str string) error {
	if !IsValidString(str, RegexpUserNick, 255) {
		return environment.ErrFormatUserNickname
	}
	return nil
}

func IsValidUserName(str string) error {
	if !IsValidString(str, RegexpUserName, 255) {
		return environment.ErrFormatUserName
	}
	return nil
}

func IsValidUserEmail(str string) error {
	if !IsValidString(str, RegexpUserEmail, 255) {
		return environment.ErrFormatUserEmail
	}
	return nil
}

func IsValidUserPhone(str string) error {
	if str == "" {
		return nil // allow no phone
	}
	newstr := regexp.MustCompile(`\s`).ReplaceAllString(str, "")
	if !IsValidString(newstr, RegexpUserPhone, 255) {
		return environment.ErrFormatUserPhone
	}
	return nil
}
