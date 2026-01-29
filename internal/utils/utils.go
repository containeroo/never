package utils

import "strings"

func IsResolvableValue(s string) bool {
	switch {
	case strings.HasPrefix(s, "env:"):
		return true
	case strings.HasPrefix(s, "file:"):
		return true
	case strings.HasPrefix(s, "json:"):
		return true
	case strings.HasPrefix(s, "yaml:"):
		return true
	case strings.HasPrefix(s, "ini:"):
		return true
	default:
		return false
	}
}
