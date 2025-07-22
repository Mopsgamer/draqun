package htmx

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrDatabase = NewHTMXAlert(
		errors.Join(fiber.ErrInternalServerError, errors.New("database: failed")),
		"Internal database error.",
		Danger,
	)
	ErrUseless = NewHTMXAlert(
		errors.Join(fiber.ErrNotFound, errors.New("change: useless")),
		"Useless change.",
		Danger,
	)

	ErrToken = NewHTMXAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid")),
		"Invalid token.",
		Danger,
	)

	ErrUserNotFound = NewHTMXAlert(
		errors.Join(fiber.ErrNotFound, errors.New("user: not found")),
		"User not found.",
		Danger,
	)
	ErrUserNickname = NewHTMXAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid nickname")),
		"Invalid nickname.",
		Danger,
	)
	ErrUserNameConfirm = NewHTMXAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation")),
		"Invalid name confirmation.",
		Danger,
	)
	ErrUserPassword = NewHTMXAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password")),
		"Invalid password.",
		Danger,
	)
	ErrUserPasswordConfirm = NewHTMXAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation")),
		"Invalid password confirmation.",
		Danger,
	)
	ErrUserExsistsEmail = NewHTMXAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists")),
		"Email already exists.",
		Danger,
	)
	ErrUserExsistsName = NewHTMXAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this username exists")),
		"Username already exists.",
		Danger,
	)
	ErrUserExsistsNickname = NewHTMXAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this nickname exists")),
		"Nickname already exists.",
		Danger,
	)
	ErrUserDeleteOwnerAccount = NewHTMXAlert(
		errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner")),
		"Cannot delete account as group owner.",
		Danger,
	)

	ErrGroupNotFound = NewHTMXAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: not found")),
		"Group not found.",
		Danger,
	)

	ErrGroupMemberNotFound = NewHTMXAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: member: not found")),
		"Group member not found.",
		Danger,
	)
	ErrGroupMemberIsOnlyAdmin = NewHTMXAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: last admin")),
		"Group member is only admin.",
		Danger,
	)
	ErrGroupMemberNotAllowed = NewHTMXAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed")),
		"Group member rights not allowed.",
		Danger,
	)

	ErrGroupChatInvalidContent = NewHTMXAlert(
		errors.Join(fiber.ErrUnprocessableEntity, errors.New("group: chat: invalid content")),
		"Invalid chat message content.",
		Danger,
	)
)

var (
	ErrFormatUserPassword = NewHTMXAlert(
		errors.New("format: user: invalid password"),
		"Invalid password format.",
		Warning,
	)
	ErrFormatUserEmail = NewHTMXAlert(
		errors.New("format: user: invalid email"),
		"Invalid email format.",
		Warning,
	)
	ErrFormatUserMoniker = NewHTMXAlert(
		errors.New("format: user: invalid moniker"),
		"Invalid moniker format.",
		Warning,
	)
	ErrFormatUserName = NewHTMXAlert(
		errors.New("format: user: invalid name"),
		"Invalid name format.",
		Warning,
	)
	ErrFormatUserPhone = NewHTMXAlert(
		errors.New("format: user: invalid phone"),
		"Invalid phone format.",
		Warning,
	)

	ErrFormatGroupName = NewHTMXAlert(
		errors.New("format: invalid group name"),
		"Invalid group name format.",
		Warning,
	)
	ErrFormatGroupMoniker = NewHTMXAlert(
		errors.New("format: invalid group moniker"),
		"Invalid group moniker format.",
		Warning,
	)
	ErrFormatGroupPassword = NewHTMXAlert(
		errors.New("format: invalid group password"),
		"Invalid group password format.",
		Warning,
	)
	ErrFormatGroupDescription = NewHTMXAlert(
		errors.New("format: invalid group description"),
		"Invalid group description format.",
		Warning,
	)
	ErrFormatGroupMode = NewHTMXAlert(
		errors.New("format: invalid group mode"),
		"Invalid group mode format.",
		Warning,
	)
)
