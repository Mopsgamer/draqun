package model_request

type UserDelete struct {
	CurrentPassword string `form:"current-password"`
	ConfirmUsername string `form:"confirm-username"`
}
