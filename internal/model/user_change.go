package model

// Supports: json, form.
type UserChangeName struct {
	NewNickname string `json:"new-nickname" form:"new-nickname"`
	NewUsername string `json:"new-username" form:"new-username"`
}

// Supports: json, form.
type UserChangeEmail struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewEmail        string `json:"new-email" form:"new-email"`
}

// Supports: json, form.
type UserChangePhone struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewPhone        string `json:"new-phone" form:"new-phone"`
}

// Supports: json, form.
type UserChangePassword struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	NewPassword     string `json:"new-password" form:"new-password"`
	ConfirmPassword string `json:"confirm-password" form:"confirm-password"`
}
