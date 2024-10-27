package internal

type LoginRequest struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (req LoginRequest) IsBad() bool {
	return req.Email == "" || req.Password == ""
}
