package htmx

import (
	"github.com/Mopsgamer/draqun/server/model"
)

const (
	RegexpGroupNick        string = RegexpUserNick
	RegexpGroupName        string = RegexpUserName
	RegexpGroupPassword    string = "(^$|" + RegexpUserPassword + ")"
	RegexpGroupDescription string = "^.{0,500}$"
)

func IsValidGroupNick(str string) error {
	if !IsValidString(str, RegexpGroupNick, 1, 255) {
		return ErrFormatGroupMoniker
	}
	return nil
}

func IsValidGroupName(str string) error {
	if !IsValidString(str, RegexpGroupName, 1, 255) {
		return ErrFormatGroupName
	}
	return nil
}

func IsValidGroupPassword(str string) error {
	if !IsValidString(str, RegexpGroupPassword, 0, 255) {
		return ErrFormatGroupPassword
	}
	return nil
}

func IsValidGroupDescription(str string) error {
	if !IsValidString(str, RegexpGroupDescription, 0, 500) {
		return ErrFormatGroupDescription
	}
	return nil
}

func IsValidGroupMode(val string) error {
	if !IsValidEnumString(val, []model.GroupMode{model.GroupModeDm, model.GroupModePrivate, model.GroupModePublic}) {
		return ErrFormatGroupMode
	}
	return nil
}
