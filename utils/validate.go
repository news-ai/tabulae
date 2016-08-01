package utils

import (
	"strings"

	"golang.org/x/net/html"
)

func ValidateHTML(validate string) bool {
	_, err := html.Parse(strings.NewReader(validate))
	if err != nil {
		return false
	}
	return true
}
