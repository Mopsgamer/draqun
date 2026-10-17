package model

import (
	"crypto/sha256"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"golang.org/x/crypto/bcrypt"
)

type Moniker string

var regexpNick = regexp.MustCompile(`^.{1,255}$`)

func (nick Moniker) IsValid() bool {
	return regexpNick.Match([]byte(nick))
}

type Name string

var regexpName = regexp.MustCompile(`^[a-zA-Z0-9._]{1,255}$`)

func (name Name) IsValid() bool {
	return regexpName.Match([]byte(name))
}

type Email string

var regexpEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (email Email) IsValid() bool {
	return regexpEmail.Match([]byte(email))
}

type Password string

func (password Password) IsValid() bool {
	return len(password) >= 8 && len(password) <= 255
}

func (password Password) Hash() (PasswordHashed, error) {
	sum := sha256.Sum256([]byte(password))
	hash, err := bcrypt.GenerateFromPassword(sum[:], bcrypt.DefaultCost)
	return PasswordHashed(hash), err
}

type PasswordHashed string

func (passwordHashed PasswordHashed) IsValid() bool {
	return len(passwordHashed) > 0
}

func (passwordHashed PasswordHashed) Compare(password Password) error {
	sum := sha256.Sum256([]byte(password))
	return bcrypt.CompareHashAndPassword([]byte(passwordHashed), sum[:])
}

type OptionalPassword Password

func (password OptionalPassword) IsValid() bool {
	return len(password) == 0 || Password(password).IsValid()
}

func (password OptionalPassword) Hash() (OptionalPasswordHashed, error) {
	if len(password) == 0 {
		return "", nil
	}
	hash, err := Password(password).Hash()
	return OptionalPasswordHashed(hash), err
}

type OptionalPasswordHashed PasswordHashed

func (password OptionalPasswordHashed) IsValid() bool {
	return true
}

func (passwordHashed OptionalPasswordHashed) Compare(password OptionalPassword) error {
	return PasswordHashed(passwordHashed).Compare(Password(password))
}

type Phone string

var regexpPhone = regexp.MustCompile(`^\+?\d+$`)

func (phone Phone) IsValid() bool {
	return len(phone) == 0 || len(phone) >= 10 && len(phone) <= 15 && regexpPhone.Match([]byte(phone))
}

type Description string

func (description Description) IsValid() bool {
	return len(description) <= 500
}

type MessageContent string

func (messageContent MessageContent) IsValid() bool {
	return len(strings.TrimSpace(string(messageContent))) > 0 && len(messageContent) <= environment.ChatMessageMaxLength
}

type Avatar string

func (avatar Avatar) IsValid() bool {
	return len(avatar) <= 255
}

type Color uint32

func (color Color) IsValid() bool {
	return color <= 0xFFFFFF
}

func (color Color) Hex() string {
	return fmt.Sprintf("#%06x", uint32(color))
}

func ParseColorHex(s string) Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 0 {
		return 0
	}
	var val uint32
	_, _ = fmt.Sscanf(s, "%x", &val)
	return Color(val)
}

type TimePast time.Time

func (tp TimePast) Value() (driver.Value, error) {
	t := time.Time(tp)
	if t.IsZero() {
		return nil, nil
	}
	return t.Format(time.RFC3339), nil
}

func (tp *TimePast) Scan(src interface{}) error {
	if src == nil {
		*tp = TimePast(time.Time{})
		return nil
	}
	var str string
	switch v := src.(type) {
	case time.Time:
		*tp = TimePast(v)
		return nil
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported Scan, storing %T into TimePast", src)
	}

	for _, fmtStr := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999+00:00",
	} {
		if t, err := time.Parse(fmtStr, str); err == nil {
			*tp = TimePast(t)
			return nil
		}
	}
	return fmt.Errorf("cannot parse time string %q into TimePast", str)
}

func (tp TimePast) IsValid() bool {
	return !time.Time(tp).IsZero()
}

type TimeFuture time.Time

func (tp TimeFuture) Value() (driver.Value, error) {
	t := time.Time(tp)
	if t.IsZero() {
		return nil, nil
	}
	return t.Format(time.RFC3339), nil
}

func (tf *TimeFuture) Scan(src interface{}) error {
	if src == nil {
		*tf = TimeFuture(time.Time{})
		return nil
	}
	var str string
	switch v := src.(type) {
	case time.Time:
		*tf = TimeFuture(v)
		return nil
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("unsupported Scan, storing %T into TimeFuture", src)
	}

	for _, fmtStr := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999+00:00",
	} {
		if t, err := time.Parse(fmtStr, str); err == nil {
			*tf = TimeFuture(t)
			return nil
		}
	}
	return fmt.Errorf("cannot parse time string %q into TimeFuture", str)
}

func (tf TimeFuture) IsValid() bool {
	return !time.Time(tf).IsZero()
}
