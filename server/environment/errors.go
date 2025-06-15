package environment

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrUseless = errors.Join(fiber.ErrBadRequest, errors.New("change: useless"))

	ErrToken = errors.Join(fiber.ErrBadRequest, errors.New("token: invalid"))

	ErrUserNickname           = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid nickname"))
	ErrUserName               = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid name"))
	ErrUserNameConfirm        = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid name confirmation"))
	ErrUserPassword           = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid password"))
	ErrUserPasswordConfirm    = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid password confirmation"))
	ErrUserEmail              = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid email"))
	ErrUserPhone              = errors.Join(fiber.ErrBadRequest, errors.New("user: invalid phone"))
	ErrUserExsistsEmail       = errors.Join(fiber.ErrBadRequest, errors.New("user: another user with this email exists"))
	ErrUserExsistsName        = errors.Join(fiber.ErrBadRequest, errors.New("user: another user with this username exists"))
	ErrUserExsistsNickname    = errors.Join(fiber.ErrBadRequest, errors.New("user: another user with this nickname exists"))
	ErrUserNotFound           = errors.Join(fiber.ErrBadRequest, errors.New("user: not found"))
	ErrUserDeleteOwnerAccount = errors.Join(fiber.ErrBadRequest, errors.New("user: cannot delete account if a group owner"))

	ErrGroupNotFound    = errors.Join(fiber.ErrBadRequest, errors.New("group: not found"))
	ErrGroupName        = errors.Join(fiber.ErrBadRequest, errors.New("group: invalid name"))
	ErrGroupNick        = errors.Join(fiber.ErrBadRequest, errors.New("group: invalid nickname"))
	ErrGroupPassword    = errors.Join(fiber.ErrBadRequest, errors.New("group: invalid password"))
	ErrGroupDescription = errors.Join(fiber.ErrBadRequest, errors.New("group: invalid description"))
	ErrGroupMode        = errors.Join(fiber.ErrBadRequest, errors.New("group: invalid mode"))

	ErrGroupMemberNotFound   = errors.Join(fiber.ErrBadRequest, errors.New("group: member: not found"))
	ErrGroupMemberNotAllowed = errors.Join(fiber.ErrBadRequest, errors.New("group: member: rights: not allowed"))

	ErrChatMessageContent = errors.Join(fiber.ErrBadRequest, errors.New("chat message: invalid content"))
)
