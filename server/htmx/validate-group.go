package htmx

import (
	"regexp"
	"slices"

	"github.com/Mopsgamer/draqun/server/model"
)

var (
	RegexpGroupNick        = RegexpUserNick
	RegexpGroupName        = RegexpUserName
	RegexpGroupPassword    = regexp.MustCompile("(^$|" + RegexpUserPassword.String() + ")")
	RegexpGroupDescription = regexp.MustCompile("^.{0,500}$")
)

func IsValidGroupNick(str string) error {
	if !RegexpGroupNick.Match([]byte(str)) {
		return ErrFormatGroupMoniker
	}
	return nil
}

func IsValidGroupName(str string) error {
	if !RegexpGroupName.Match([]byte(str)) {
		return ErrFormatGroupName
	}
	return nil
}

func IsValidGroupPassword(str string) error {
	if !RegexpGroupPassword.Match([]byte(str)) {
		return ErrFormatGroupPassword
	}
	return nil
}

func IsValidGroupDescription(str string) error {
	if !RegexpGroupDescription.Match([]byte(str)) {
		return ErrFormatGroupDescription
	}
	return nil
}

func IsValidGroupMode(val string) error {
	if !slices.Contains([]model.GroupMode{model.GroupModeDm, model.GroupModePrivate, model.GroupModePublic}, model.GroupMode(val)) {
		return ErrFormatGroupMode
	}
	return nil
}
