package model_database

import "errors"

var (
	ErrFormatGroupName        = errors.New("format: invalid group name")
	ErrFormatGroupNick        = errors.New("format: invalid group nickname")
	ErrFormatGroupPassword    = errors.New("format: invalid group password")
	ErrFormatGroupDescription = errors.New("format: invalid group description")
	ErrFormatGroupMode        = errors.New("format: invalid group mode")
)

const (
	RegexpGroupNick        string = RegexpUserNick
	RegexpGroupName        string = RegexpUserName
	RegexpGroupPassword    string = "(^$|" + RegexpUserPassword + ")"
	RegexpGroupDescription string = "^.{0,500}$"
)

func IsValidGroupNick(str string) error {
	if !IsValidString(str, RegexpGroupNick, 255) {
		return ErrFormatGroupNick
	}
	return nil
}

func IsValidGroupName(str string) error {
	if !IsValidString(str, RegexpGroupName, 255) {
		return ErrFormatGroupName
	}
	return nil
}

func IsValidGroupPassword(str *string) error {
	if str == nil {
		return nil // allow no group password
	}
	if !IsValidString(*str, RegexpGroupPassword, 255) {
		return ErrFormatGroupPassword
	}
	return nil
}

func IsValidGroupDescription(str string) error {
	if !IsValidString(str, RegexpGroupDescription, 500) {
		return ErrFormatGroupDescription
	}
	return nil
}

func IsValidGroupMode(val string) error {
	if !IsValidEnumString(val, []GroupMode{GroupModeDm, GroupModePrivate, GroupModePublic}) {
		return ErrFormatGroupMode
	}
	return nil
}
