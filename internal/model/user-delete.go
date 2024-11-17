package model

// Supports: json, form.
type UserDelete struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	ConfirmUsername string `json:"confirm-username" form:"confirm-username"`
}
