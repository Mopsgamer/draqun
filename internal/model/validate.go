package model

import (
	"regexp"
)

func ValidateString(str string, rg string) bool {
	l := len(str)
	return l >= 0 && l <= 255 && regexp.MustCompile(rg).MatchString(str)
}
