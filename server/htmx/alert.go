package htmx

import (
	"errors"
	"fmt"
)

type AlertLevel int

var _ fmt.Stringer = (*AlertLevel)(nil)
var _ fmt.GoStringer = (*AlertLevel)(nil)

func (level AlertLevel) String() string {
	switch level {
	case Primary:
		return "primary"
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Danger:
		return "danger"
	default:
		panic(fmt.Sprintf("Unknown alert level: %d", level))
	}
}

func (level AlertLevel) GoString() string {
	return level.String()
}

const (
	Primary AlertLevel = iota
	Success
	Warning
	Danger
)

type Alert interface {
	error
	Local() string // User friendly error message.
	Level() AlertLevel
}

type alert struct {
	err   error
	local string // User friendly error message.
	level AlertLevel
}

var _ Alert = (*alert)(nil)

func NewAlert(err error, local string, level AlertLevel) alert {
	return alert{
		err:   err,
		local: local,
		level: level,
	}
}

func (a alert) Join(errs ...error) alert {
	errs = append([]error{a.err}, errs...)
	a.err = errors.Join(errs...)
	return a
}

func (a alert) Is(err error) bool {
	return errors.Is(a.err, err)
}

func (a alert) Error() string {
	return a.err.Error()
}

// User friendly error message.
func (a alert) Local() string {
	return a.local
}

func (a alert) Level() AlertLevel {
	return a.level
}
