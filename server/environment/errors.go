package environment

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrUseless = errors.Join(fiber.ErrNotFound, errors.New("change: useless"))

	ErrToken = errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid"))

	ErrUserNotFound           = errors.Join(fiber.ErrNotFound, errors.New("user: not found"))
	ErrUserNickname           = errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid nickname"))
	ErrUserNameConfirm        = errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation"))
	ErrUserPassword           = errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password"))
	ErrUserPasswordConfirm    = errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation"))
	ErrUserExsistsEmail       = errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists"))
	ErrUserExsistsName        = errors.Join(fiber.ErrConflict, errors.New("user: another user with this username exists"))
	ErrUserExsistsNickname    = errors.Join(fiber.ErrConflict, errors.New("user: another user with this nickname exists"))
	ErrUserDeleteOwnerAccount = errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner"))

	ErrGroupNotFound = errors.Join(fiber.ErrNotFound, errors.New("group: not found"))

	ErrGroupMemberNotFound   = errors.Join(fiber.ErrNotFound, errors.New("group: member: not found"))
	ErrGroupMemberNotAllowed = errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed"))

	ErrChatMessageContent = errors.Join(fiber.ErrUnprocessableEntity, errors.New("chat message: invalid content"))
)
