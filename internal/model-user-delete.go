package internal

// Supports: json, form.
type UserDelete struct {
	CurrentPassword string `json:"current-password" form:"current-password"`
	ConfirmName     string `json:"confirm-name" form:"confirm-name"`
}

// Checks if the request contains invalid current password or confirm name fields.
func (req UserDelete) IsBad() bool {
	return req.CurrentPassword == "" || req.ConfirmName == ""
}
