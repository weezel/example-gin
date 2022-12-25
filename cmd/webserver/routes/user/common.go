package user

import (
	"unicode"
	"unicode/utf8"
)

func isValidName(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}

	for _, c := range s {
		if unicode.IsControl(c) || unicode.IsPunct(c) {
			return false
		}
	}
	return true
}
