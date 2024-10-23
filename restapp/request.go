package restapp

type RegisterRequest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
