package database

type Action interface {
	Kind() string
}
