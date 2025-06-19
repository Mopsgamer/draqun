package environment

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrUseless = NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("change: useless")), "Useless change.", Danger)

	ErrToken = NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid")), "Invalid token.", Danger)

	ErrUserNotFound           = NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("user: not found")), "User not found.", Danger)
	ErrUserNickname           = NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid nickname")), "Invalid nickname.", Danger)
	ErrUserNameConfirm        = NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation")), "Invalid name confirmation.", Danger)
	ErrUserPassword           = NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password")), "Invalid password.", Danger)
	ErrUserPasswordConfirm    = NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation")), "Invalid password confirmation.", Danger)
	ErrUserExsistsEmail       = NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists")), "Email already exists.", Danger)
	ErrUserExsistsName        = NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this username exists")), "Username already exists.", Danger)
	ErrUserExsistsNickname    = NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this nickname exists")), "Nickname already exists.", Danger)
	ErrUserDeleteOwnerAccount = NewHTMXAlert(errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner")), "Cannot delete account as group owner.", Danger)

	ErrGroupNotFound = NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("group: not found")), "Group not found.", Danger)

	ErrGroupMemberNotFound   = NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("group: member: not found")), "Group member not found.", Danger)
	ErrGroupMemberNotAllowed = NewHTMXAlert(errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed")), "Group member rights not allowed.", Danger)

	ErrChatMessageContent = NewHTMXAlert(errors.Join(fiber.ErrUnprocessableEntity, errors.New("chat message: invalid content")), "Invalid chat message content.", Danger)
)

var (
	ErrFormatUserPassword = errors.New("format: user: invalid password")
	ErrFormatUserEmail    = errors.New("format: user: invalid email")
	ErrFormatUserNickname = errors.New("format: user: invalid nickname")
	ErrFormatUserName     = errors.New("format: user: invalid name")
	ErrFormatUserPhone    = errors.New("format: user: invalid phone")

	ErrFormatGroupName        = errors.New("format: invalid group name")
	ErrFormatGroupNick        = errors.New("format: invalid group nickname")
	ErrFormatGroupPassword    = errors.New("format: invalid group password")
	ErrFormatGroupDescription = errors.New("format: invalid group description")
	ErrFormatGroupMode        = errors.New("format: invalid group mode")
)
