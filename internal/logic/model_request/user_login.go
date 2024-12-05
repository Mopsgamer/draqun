package model_request

type UserLogin struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}
