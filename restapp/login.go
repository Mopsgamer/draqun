package restapp

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginRequest) IsBad() bool {
	return req.Email == "" || req.Password == ""
}
