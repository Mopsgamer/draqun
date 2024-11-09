package internal

// Supports: json, form.
type UserLogin struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// Checks if the request contains invalid email and password fields.
func (req UserLogin) IsBad() bool {
	return req.Email == "" || req.Password == ""
}
