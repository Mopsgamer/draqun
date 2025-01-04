package model_http

type UserLogin struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}
