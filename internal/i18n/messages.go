package i18n

import "restapp/internal/logic/model_database"

const (
	MessageFatalDatabaseQuery   = "Fatal database error."
	MessageFatalTokenGeneration = "Fatal token generation."

	MessageDetailGroupNick        = MessageDetailUserNick
	MessageDetailGroupName        = MessageDetailUserName
	MessageDetailGroupPassword    = MessageDetailUserPassword
	MessageDetailGroupDescription = "Must be between 0 and 500 characters long and can contain any characters."
	MessageDetailGroupMode        = "Must be " + model_database.GroupModeDm + ", " + model_database.GroupModePrivate + " or " + model_database.GroupModePrivate + "."
	MessageDetailUserName         = "Must contain only letters (A-Z, a-z), numbers (0-9), and these special characters: . _ . No spaces. Must be at least 1 characters long and no more than 255 characters."
	MessageDetailUserPassword     = "Must contain only letters (A-Z, a-z), numbers (0-9), spaces, or these special characters: , . ~ - + % $ ^ & * _ ! ? ( ) [ ] { } `. Must be at least 8 characters long and no more than 255 characters."
	MessageDetailUserNick         = "Must be between 1 and 255 characters long and can contain any characters."
	MessageDetailUserEmail        = "Must be a valid email."
	MessageDetailUSerPhone        = "Must be a valid phone number."

	MessageErrNoRights                      = "You do not have the necessary rights or permissions to perform this action."
	MessageErrGroupNotFound                 = "Group not found."
	MessageErrGroupExistsGroupname          = "This group name is taken."
	MessageErrGroupName                     = "Invalid group name pattern. " + MessageDetailGroupName
	MessageErrGroupNick                     = "Invalid group nick name pattern. " + MessageDetailGroupNick
	MessageErrGroupPassword                 = "Invalid group password pattern. You can leave this field empty. " + MessageDetailGroupPassword
	MessageErrGroupDescription              = "Invalid group description. " + MessageDetailGroupDescription
	MessageErrGroupMode                     = "Invalid group mode. " + MessageDetailGroupMode
	MessageErrUselessChange                 = "No changes. "
	MessageErrMessageContent                = "The message must be between 1 and " + model_database.ContentMaxLengthString + " characters long."
	MessageErrNotGroupMember                = "Not a member of the group."
	MessageErrAlreadyGroupMember            = "Already a member of the group."
	MessageErrInvalidRequest                = "Invalid request payload."
	MessageErrPassword                      = "Invalid password pattern. " + MessageDetailUserPassword
	MessageErrPasswordSame                  = "The new password is the same as the old one."
	MessageErrUserNick                      = "Invalid nick name pattern. " + MessageDetailUserNick
	MessageErrUserName                      = "Invalid user name pattern. " + MessageDetailUserName
	MessageErrEmail                         = "Invalid email pattern. " + MessageDetailUserEmail
	MessageErrEmailSame                     = "The new email is the same as the old one."
	MessageErrPhone                         = "Invalid phone number pattern. " + MessageDetailUSerPhone
	MessageErrPhoneSame                     = "The new phone is the same as the old one."
	MessageErrBadConfirmPassword            = "Passwords are not same."
	MessageErrBadUsernameConfirm            = "Usernames are not same."
	MessageErrBadPassword                   = "Invalid user password."
	MessageErrUserNotFound                  = "User not found."
	MessageErrUserExistsUsername            = "This user name is taken."
	MessageErrUserExistsEmail               = "This email is taken."
	MessageErrUserExistsPhone               = "This phone number is taken."
	MessageErrCanNotDeleteGroupOwnerAccount = "The user cannot be deleted because the user is the owner of a group or set of groups."

	MessageSuccChangedProfile = "Successfully changed the user profile."
	MessageSuccChangedPass    = "Successfully changed the user password."
	MessageSuccChangedEmail   = "Successfully changed the user email."
	MessageSuccChangedPhone   = "Successfully changed the user phone."
	MessageSuccDeletedUser    = "Successfully deleted the user."
	MessageSuccLeavedGroup    = "Successfully leaved from the group."
	MessageSuccJoinedGroup    = "Successfully join to the group."
	MessageSuccCreatedGroup   = "Successfully created the group."
	MessageSuccLogin          = "Successfully logged in! Redirecting..."
)
