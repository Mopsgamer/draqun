package database

import "github.com/Mopsgamer/draqun/server/environment"

const (
	RegexpGroupNick        string = RegexpUserNick
	RegexpGroupName        string = RegexpUserName
	RegexpGroupPassword    string = "(^$|" + RegexpUserPassword + ")"
	RegexpGroupDescription string = "^.{0,500}$"
)

func IsValidGroupNick(str string) error {
	if !IsValidString(str, RegexpGroupNick, 255) {
		return environment.ErrFormatGroupNick
	}
	return nil
}

func IsValidGroupName(str string) error {
	if !IsValidString(str, RegexpGroupName, 255) {
		return environment.ErrFormatGroupName
	}
	return nil
}

func IsValidGroupPassword(str string) error {
	if str == "" {
		return nil // allow no group password
	}
	if !IsValidString(str, RegexpGroupPassword, 255) {
		return environment.ErrFormatGroupPassword
	}
	return nil
}

func IsValidGroupDescription(str string) error {
	if !IsValidString(str, RegexpGroupDescription, 500) {
		return environment.ErrFormatGroupDescription
	}
	return nil
}

func IsValidGroupMode(val string) error {
	if !IsValidEnumString(val, []GroupMode{GroupModeDm, GroupModePrivate, GroupModePublic}) {
		return environment.ErrFormatGroupMode
	}
	return nil
}
