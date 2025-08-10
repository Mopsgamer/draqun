package htmx

import (
	"errors"
	"fmt"
)

type ShoelaceAlertLevel int

var _ fmt.Stringer = (*ShoelaceAlertLevel)(nil)
var _ fmt.GoStringer = (*ShoelaceAlertLevel)(nil)

func (level ShoelaceAlertLevel) String() string {
	switch level {
	case Success:
		return "success"
	case Warning:
		return "warning"
	case Danger:
		return "danger"
	default:
		panic(fmt.Sprintf("Unknown shoelace alert level: %d", level))
	}
}

func (level ShoelaceAlertLevel) GoString() string {
	return level.String()
}

const (
	Primary ShoelaceAlertLevel = iota
	Success
	Warning
	Danger
)

type Alert interface {
	error
	Local() string // User friendly error message.
	Level() ShoelaceAlertLevel
}

type alert struct {
	err   error
	local string // User friendly error message.
	level ShoelaceAlertLevel
}

var _ Alert = (*alert)(nil)

func NewAlert(err error, local string, level ShoelaceAlertLevel) alert {
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

func (a alert) Level() ShoelaceAlertLevel {
	return a.level
}
