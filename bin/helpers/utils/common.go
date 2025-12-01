package utils

import (
	"strings"
)

func ToSlug(s string) string {
	s = strings.TrimSpace(s)                 // remove leading/trailing spaces
	s = strings.ToLower(s)                   // lowercase
	s = strings.Join(strings.Fields(s), "-") // replace multiple spaces with '-'
	return s
}
