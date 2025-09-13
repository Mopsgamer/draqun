package model

import (
	"regexp"
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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return PasswordHashed(hash), err
}

type PasswordHashed string

func (passwordHashed PasswordHashed) IsValid() bool {
	return len(passwordHashed) > 0
}

func (passwordHashed PasswordHashed) Compare(password Password) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(password))
}

type OptionalPassword Password

func (password OptionalPassword) IsValid() bool {
	return len(password) == 0 || Password(password).IsValid()
}

func (password OptionalPassword) Hash() (OptionalPasswordHashed, error) {
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
	return len(messageContent) <= environment.ChatMessageMaxLength
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

func (tp TimePast) IsValid() bool {
	now := time.Now()
	return time.Time(tp).Before(now) || now.Equal(time.Time(tp))
}

type TimeFuture time.Time

func (tf TimeFuture) IsValid() bool {
	now := time.Now()
	return time.Time(tf).After(now) || now.Equal(time.Time(tf))
}
