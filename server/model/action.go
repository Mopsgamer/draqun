package model

type Action interface {
	Model
	SetDb(db *DB)
	Kind() string
}
