package model

import "github.com/Mopsgamer/draqun/server/htmx"

type Model interface {
	Insert() error
	Update() error

	IsEmpty() bool
	Validate() htmx.Alert
}

type SetFullfilled interface {
	SetFullfilled()
}

type checkEmpty struct {
	fullfilled bool `db:"-"`
}

var _ SetFullfilled = (*checkEmpty)(nil)

func (check checkEmpty) IsEmpty() bool {
	return !check.fullfilled
}

func (check *checkEmpty) SetFullfilled() {
	check.fullfilled = true
}
