package model_database

import (
	"regexp"
	"slices"
)

func IsValidString(str string, rg string, maxLen int) bool {
	l := len(str)
	return l >= 0 && l <= maxLen && regexp.MustCompile(rg).MatchString(str)
}

func IsValidEnum(str string, values []string) bool {
	return slices.Contains(values, str)
}
