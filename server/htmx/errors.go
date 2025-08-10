package htmx

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	AlertEncryption = NewAlert(
		errors.Join(fiber.ErrInternalServerError, errors.New("encryption: failed")),
		"Internal encryption error.",
		Danger,
	)
	AlertDatabase = NewAlert(
		errors.Join(fiber.ErrInternalServerError, errors.New("database: failed")),
		"Internal database error.",
		Danger,
	)
	AlertUseless = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("change: useless")),
		"Useless change.",
		Danger,
	)

	ErrToken   = errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid"))
	AlertToken = NewAlert(
		ErrToken,
		"Invalid token.",
		Danger,
	)

	AlertUserUnauthorized = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: unauthorized")),
		"Unauthorized user.",
		Danger,
	)
	AlertUserNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("user: not found")),
		"User not found.",
		Danger,
	)
	AlertUserNickname = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid nickname")),
		"Invalid nickname.",
		Danger,
	)
	AlertUserNameConfirm = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation")),
		"Invalid name confirmation.",
		Danger,
	)
	AlertUserPassword = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password")),
		"Invalid password.",
		Danger,
	)
	AlertUserPasswordConfirm = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation")),
		"Invalid password confirmation.",
		Danger,
	)
	AlertUserExsistsEmail = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists")),
		"Email already exists.",
		Danger,
	)
	AlertUserExsistsName = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this username exists")),
		"Username already exists.",
		Danger,
	)
	AlertUserExsistsNickname = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this nickname exists")),
		"Nickname already exists.",
		Danger,
	)
	AlertUserDeleteOwnerAccount = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner")),
		"Cannot delete account as group owner.",
		Danger,
	)

	AlertGroupNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: not found")),
		"Group not found.",
		Danger,
	)
	AlertGroupExsistsName = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: another group with this name exists")),
		"Group with this name already exists.",
		Danger,
	)

	AlertGroupMemberNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: member: not found")),
		"Group member not found.",
		Danger,
	)
	AlertGroupMemberIsOnlyOwner = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: last owner")),
		"Group member is only owner.",
		Danger,
	)
	AlertGroupMemberNotAllowed = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed")),
		"Group member rights not allowed.",
		Danger,
	)
)

var (
	ErrFormat = errors.Join(fiber.ErrUnprocessableEntity, errors.New("invalid data format"))

	AlertFormatMoniker = NewAlert(
		errors.Join(ErrFormat, errors.New("bad moniker")),
		"Invalid moniker format.",
		Warning,
	)
	AlertFormatName = NewAlert(
		errors.Join(ErrFormat, errors.New("bad name")),
		"Invalid name format.",
		Warning,
	)
	AlertFormatEmail = NewAlert(
		errors.Join(ErrFormat, errors.New("bad email")),
		"Invalid email format.",
		Warning,
	)
	AlertFormatPassword = NewAlert(
		errors.Join(ErrFormat, errors.New("bad password")),
		"Invalid password format.",
		Warning,
	)
	AlertFormatPhone = NewAlert(
		errors.Join(ErrFormat, errors.New("bad phone")),
		"Invalid phone format.",
		Warning,
	)
	AlertFormatDescription = NewAlert(
		errors.Join(ErrFormat, errors.New("bad description")),
		"Invalid description format.",
		Warning,
	)
	AlertFormatMessageContent = NewAlert(
		errors.Join(ErrFormat, errors.New("bad message content")),
		"Invalid message format.",
		Warning,
	)
	AlertFormatAvatar = NewAlert(
		errors.Join(ErrFormat, errors.New("bad avatar")),
		"Invalid avatar format.",
		Warning,
	)
	AlertFormatColor = NewAlert(
		errors.Join(ErrFormat, errors.New("bad color")),
		"Invalid color.",
		Warning,
	)
	AlertFormatPastMoment = NewAlert(
		errors.Join(ErrFormat, errors.New("bad past moment")),
		"Invalid past time moment.",
		Warning,
	)
	AlertFormatFutureMoment = NewAlert(
		errors.Join(ErrFormat, errors.New("bad future moment")),
		"Invalid future time moment.",
		Warning,
	)

	AlertFormatGroupMode = NewAlert(
		errors.Join(ErrFormat, errors.New("bad group mode")),
		"Invalid group mode format.",
		Warning,
	)
	ErrFormatGroupPerm         = errors.Join(ErrFormat, errors.New("bad group permission"))
	AlertFormatGroupPermSwitch = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group message permission")),
		"Invalid message permission format.",
		Warning,
	)
	AlertFormatGroupPermMessages = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group message permission")),
		"Invalid message permission format.",
		Warning,
	)
	AlertFormatGroupPermMembers = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group member permission")),
		"Invalid member permission format.",
		Warning,
	)
)
