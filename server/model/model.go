package model

import "github.com/Mopsgamer/draqun/server/htmx"

type Model interface {
	Insert() error
	Update() error

	IsEmpty() bool
	IsValid() htmx.Alert
}
