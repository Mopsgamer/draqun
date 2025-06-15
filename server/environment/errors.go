package environment

import "errors"

var (
	ErrUseless = errors.New("change: useless")

	ErrToken = errors.New("token: invalid")

	ErrUserNickname           = errors.New("user: invalid nickname")
	ErrUserName               = errors.New("user: invalid name")
	ErrUserNameConfirm        = errors.New("user: invalid name confirmation")
	ErrUserPassword           = errors.New("user: invalid password")
	ErrUserPasswordConfirm    = errors.New("user: invalid password confirmation")
	ErrUserEmail              = errors.New("user: invalid email")
	ErrUserPhone              = errors.New("user: invalid phone")
	ErrUserExsistsEmail       = errors.New("user with this email exists")
	ErrUserExsistsName        = errors.New("user with this username exists")
	ErrUserExsistsNickname    = errors.New("user with this nickname exists")
	ErrUserNotFound           = errors.New("user: not found")
	ErrUserDeleteOwnerAccount = errors.New("user: cannot delete account if a group owner")

	ErrGroupNotFound    = errors.New("group: not found")
	ErrGroupName        = errors.New("group: invalid name")
	ErrGroupNick        = errors.New("group: invalid nickname")
	ErrGroupPassword    = errors.New("group: invalid password")
	ErrGroupDescription = errors.New("group: invalid description")
	ErrGroupMode        = errors.New("group: invalid mode")

	ErrGroupMemberNotFound   = errors.New("group: member: not found")
	ErrGroupMemberNotAllowed = errors.New("group: member: rights: not allowed")

	ErrChatMessageContent = errors.New("chat message: invalid content")
)
