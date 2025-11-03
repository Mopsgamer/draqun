package model

import (
	"crypto/sha256"
	"database/sql/driver"
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

type TimePast time.Time

func (tp TimePast) Value() (driver.Value, error) {
	t := time.Time(tp)
	if t.IsZero() {
		// return nil if you want SQL NULL for zero time
		return nil, nil
	}
	return t, nil
}

func (tp TimePast) IsValid() bool {
	now := time.Now()
	return time.Time(tp).Before(now) || now.Equal(time.Time(tp))
}

type TimeFuture time.Time

func (tp TimeFuture) Value() (driver.Value, error) {
	t := time.Time(tp)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}

func (tf TimeFuture) IsValid() bool {
	now := time.Now()
	return time.Time(tf).After(now) || now.Equal(time.Time(tf))
}
