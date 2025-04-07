package model_database

const (
	RegexpGroupNick        string = RegexpUserNick
	RegexpGroupName        string = RegexpUserName
	RegexpGroupPassword    string = "(^$|" + RegexpUserPassword + ")"
	RegexpGroupDescription string = "^.{0,500}$"
)

func IsValidGroupNick(str string) bool {
	return IsValidString(str, RegexpUserNick, 255)
}

func IsValidGroupName(str string) bool {
	return IsValidString(str, RegexpUserName, 255)
}

func IsValidGroupPassword(str *string) bool {
	if str == nil {
		return true // allow no group password
	}
	return IsValidString(*str, RegexpGroupPassword, 255)
}

func IsValidGroupDescription(str string) bool {
	return IsValidString(str, RegexpGroupDescription, 500)
}

func IsValidGroupMode(val int) bool {
	return IsValidEnum(val, []GroupMode{GroupModeDm, GroupModePrivate, GroupModePublic})
}
