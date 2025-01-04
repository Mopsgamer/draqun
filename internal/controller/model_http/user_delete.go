package model_http

type UserDelete struct {
	CurrentPassword string `form:"current-password"`
	ConfirmUsername string `form:"confirm-username"`
}
