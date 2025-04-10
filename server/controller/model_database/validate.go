package model_database

import (
	"regexp"
	"slices"
)

func IsValidString(str string, rg string, maxLen int) bool {
	l := len(str)
	return l >= 0 && l <= maxLen && regexp.MustCompile(rg).MatchString(str)
}

func IsValidEnum[T ~int](val int, choices []T) bool {
	return slices.Contains(choices, T(val))
}

func IsValidEnumString[T ~string](val string, choices []T) bool {
	return slices.Contains(choices, T(val))
}
