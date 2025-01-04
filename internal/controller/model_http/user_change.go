package model_http

type UserChangeName struct {
	NewNickname string `form:"new-nickname"`
	NewUsername string `form:"new-username"`
}

type UserChangeEmail struct {
	CurrentPassword string `form:"current-password"`
	NewEmail        string `form:"new-email"`
}

type UserChangePhone struct {
	CurrentPassword string  `form:"current-password"`
	NewPhone        *string `form:"new-phone"`
}

type UserChangePassword struct {
	CurrentPassword string `form:"current-password"`
	NewPassword     string `form:"new-password"`
	ConfirmPassword string `form:"confirm-password"`
}
