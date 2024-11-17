package model

// Supports: json, form.
type UserDelete struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	ConfirmUsername string `json:"confirm-username" form:"confirm-username"`
}

// Checks if the request contains invalid current password or confirm username fields.
func (req UserDelete) IsBad() bool {
	return req.CurrentPassword == "" || req.ConfirmUsername == ""
}
