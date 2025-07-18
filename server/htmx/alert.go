package htmx

import "fmt"

type ShoelaceAlertLevel int

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

const (
	Primary ShoelaceAlertLevel = iota
	Success
	Warning
	Danger
)

type HTMXAlert interface {
	error
	Local() string // User friendly error message.
	Level() ShoelaceAlertLevel
}

type alert struct {
	err   error
	local string // User friendly error message.
	level ShoelaceAlertLevel
}

func NewHTMXAlert(err error, local string, level ShoelaceAlertLevel) HTMXAlert {
	return &alert{
		err:   err,
		local: local,
		level: level,
	}
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
