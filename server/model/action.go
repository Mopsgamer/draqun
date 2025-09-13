package model

type Action interface {
	Model
	Kind() string
}
