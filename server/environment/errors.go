package environment

import (
	"errors"

	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/gofiber/fiber/v3"
)

var (
	ErrUseless = htmx.NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("change: useless")), "Useless change.", htmx.Danger)

	ErrToken = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid")), "Invalid token.", htmx.Danger)

	ErrUserNotFound           = htmx.NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("user: not found")), "User not found.", htmx.Danger)
	ErrUserNickname           = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid nickname")), "Invalid nickname.", htmx.Danger)
	ErrUserNameConfirm        = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation")), "Invalid name confirmation.", htmx.Danger)
	ErrUserPassword           = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password")), "Invalid password.", htmx.Danger)
	ErrUserPasswordConfirm    = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation")), "Invalid password confirmation.", htmx.Danger)
	ErrUserExsistsEmail       = htmx.NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists")), "Email already exists.", htmx.Danger)
	ErrUserExsistsName        = htmx.NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this username exists")), "Username already exists.", htmx.Danger)
	ErrUserExsistsNickname    = htmx.NewHTMXAlert(errors.Join(fiber.ErrConflict, errors.New("user: another user with this nickname exists")), "Nickname already exists.", htmx.Danger)
	ErrUserDeleteOwnerAccount = htmx.NewHTMXAlert(errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner")), "Cannot delete account as group owner.", htmx.Danger)

	ErrGroupNotFound = htmx.NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("group: not found")), "Group not found.", htmx.Danger)

	ErrGroupMemberNotFound   = htmx.NewHTMXAlert(errors.Join(fiber.ErrNotFound, errors.New("group: member: not found")), "Group member not found.", htmx.Danger)
	ErrGroupMemberNotAllowed = htmx.NewHTMXAlert(errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed")), "Group member rights not allowed.", htmx.Danger)

	ErrChatMessageContent = htmx.NewHTMXAlert(errors.Join(fiber.ErrUnprocessableEntity, errors.New("chat message: invalid content")), "Invalid chat message content.", htmx.Danger)
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
