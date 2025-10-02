package htmx

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	AlertEncryption = NewAlert(
		errors.Join(fiber.ErrInternalServerError, errors.New("encryption: failed")),
		"An internal encryption error occurred.",
		Danger,
	)
	AlertDatabase = NewAlert(
		errors.Join(fiber.ErrInternalServerError, errors.New("database: failed")),
		"An internal database error occurred.",
		Danger,
	)
	AlertUseless = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("change: useless")),
		"Change is not necessary.",
		Danger,
	)

	ErrToken   = errors.Join(fiber.ErrUnauthorized, errors.New("token: invalid"))
	AlertToken = NewAlert(
		ErrToken,
		"The token is invalid.",
		Danger,
	)

	AlertUserUnauthorized = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: unauthorized")),
		"User is unauthorized.",
		Danger,
	)
	AlertUserNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("user: not found")),
		"The user was not found.",
		Danger,
	)
	AlertUserNickname = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid moniker")),
		"The moniker is invalid.",
		Danger,
	)
	AlertUserNameConfirm = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid name confirmation")),
		"Name confirmation is invalid.",
		Danger,
	)
	AlertUserPassword = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password")),
		"The password is invalid.",
		Danger,
	)
	AlertUserPasswordConfirm = NewAlert(
		errors.Join(fiber.ErrUnauthorized, errors.New("user: invalid password confirmation")),
		"Password confirmation is invalid.",
		Danger,
	)
	AlertUserExistsEmail = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this email exists")),
		"This email already exists.",
		Danger,
	)
	AlertUserExistsName = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this name exists")),
		"This name already exists.",
		Danger,
	)
	AlertUserExistsNickname = NewAlert(
		errors.Join(fiber.ErrConflict, errors.New("user: another user with this moniker exists")),
		"This moniker already exists.",
		Danger,
	)
	AlertUserDeleteOwnerAccount = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("user: cannot delete account if a group owner")),
		"Cannot delete account while you are a group owner.",
		Danger,
	)

	AlertGroupNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: not found")),
		"The group was not found.",
		Danger,
	)
	AlertGroupExistsName = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: another group with this name exists")),
		"A group with this name already exists.",
		Danger,
	)

	AlertGroupMemberNotFound = NewAlert(
		errors.Join(fiber.ErrNotFound, errors.New("group: member: not found")),
		"The group member was not found.",
		Danger,
	)
	AlertGroupMemberIsOnlyOwner = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: last owner")),
		"The group member is the only owner.",
		Danger,
	)
	AlertGroupMemberNotAllowed = NewAlert(
		errors.Join(fiber.ErrForbidden, errors.New("group: member: rights: not allowed")),
		"The group member does not have the required rights.",
		Danger,
	)
)

var (
	ErrFormat = errors.Join(fiber.ErrUnprocessableEntity, errors.New("invalid data format"))

	AlertFormatMoniker = NewAlert(
		errors.Join(ErrFormat, errors.New("bad moniker")),
		"The moniker format is invalid.",
		Warning,
	)
	AlertFormatName = NewAlert(
		errors.Join(ErrFormat, errors.New("bad name")),
		"The name format is invalid.",
		Warning,
	)
	AlertFormatEmail = NewAlert(
		errors.Join(ErrFormat, errors.New("bad email")),
		"The email format is invalid.",
		Warning,
	)
	AlertFormatPassword = NewAlert(
		errors.Join(ErrFormat, errors.New("bad password")),
		"The password format is invalid.",
		Warning,
	)
	AlertFormatPhone = NewAlert(
		errors.Join(ErrFormat, errors.New("bad phone")),
		"The phone format is invalid.",
		Warning,
	)
	AlertFormatDescription = NewAlert(
		errors.Join(ErrFormat, errors.New("bad description")),
		"The description format is invalid.",
		Warning,
	)
	AlertFormatMessageContent = NewAlert(
		errors.Join(ErrFormat, errors.New("bad message content")),
		"The message format is invalid.",
		Warning,
	)
	AlertFormatAvatar = NewAlert(
		errors.Join(ErrFormat, errors.New("bad avatar")),
		"The avatar format is invalid.",
		Warning,
	)
	AlertFormatColor = NewAlert(
		errors.Join(ErrFormat, errors.New("bad color")),
		"The color is invalid.",
		Warning,
	)
	AlertFormatPastMoment = NewAlert(
		errors.Join(ErrFormat, errors.New("bad past moment")),
		"The past time moment is invalid.",
		Warning,
	)
	AlertFormatFutureMoment = NewAlert(
		errors.Join(ErrFormat, errors.New("bad future moment")),
		"The future time moment is invalid.",
		Warning,
	)

	AlertFormatGroupMode = NewAlert(
		errors.Join(ErrFormat, errors.New("bad group mode")),
		"The group mode format is invalid.",
		Warning,
	)
	ErrFormatGroupPerm         = errors.Join(ErrFormat, errors.New("bad group permission"))
	AlertFormatGroupPermSwitch = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group message permission")),
		"The message permission format is invalid.",
		Warning,
	)
	AlertFormatGroupPermMessages = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group message permission")),
		"The message permission format is invalid.",
		Warning,
	)
	AlertFormatGroupPermMembers = NewAlert(
		errors.Join(ErrFormatGroupPerm, errors.New("bad group member permission")),
		"The member permission format is invalid.",
		Warning,
	)
)
