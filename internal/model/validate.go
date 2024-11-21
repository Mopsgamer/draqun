package model

import (
	"regexp"
)

func IsValidString(str string, rg string, maxLen int) bool {
	l := len(str)
	return l >= 0 && l <= maxLen && regexp.MustCompile(rg).MatchString(str)
}
