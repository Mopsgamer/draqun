package restapp

import "time"

type RegisterRequest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func (req RegisterRequest) IsBad() bool {
	return req.Name == "" || req.Email == "" || req.Password == ""
}

func (req RegisterRequest) CreateUser() (User, error) {
	hash, err := HashPassword(req.Password)
	return User{
		Name:      req.Name,
		Tag:       req.Tag,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hash,
		CreatedAt: time.Now(),
	}, err
}
