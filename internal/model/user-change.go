package model

// Supports: json, form.
type UserChangeName struct {
	NewNickname string `json:"new-nickname" form:"new-nickname"`
	NewUsername string `json:"new-username" form:"new-username"`
}

// Checks if the request contains invalid new nickname or new username fields.
func (req UserChangeName) IsBad() bool {
	return req.NewNickname == "" || req.NewUsername == ""
}

// Supports: json, form.
type UserChangeEmail struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewEmail        string `json:"new-email" form:"new-email"`
}

// Checks if the request contains invalid new email or current password fields.
func (req UserChangeEmail) IsBad() bool {
	return req.CurrentPassword == "" || req.NewEmail == ""
}

// Supports: json, form.
type UserChangePhone struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewPhone        string `json:"new-phone" form:"new-phone"`
}

// Checks if the request contains invalid new phone or current password fields.
func (req UserChangePhone) IsBad() bool {
	return req.CurrentPassword == "" || req.NewPhone == ""
}

// Supports: json, form.
type UserChangePassword struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewPassword     string `json:"new-password" form:"new-password"`
	ConfirmPassword string `json:"confirm-password" form:"confirm-password"`
}

// Checks if the request contains invalid new password, confirm password or current password fields.
func (req UserChangePassword) IsBad() bool {
	return req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == ""
}
