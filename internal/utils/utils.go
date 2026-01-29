package utils

import "strings"

// IsResolvableValue returns true if the string is a resolvable value.
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

// IsHostnameLike returns true if the string is a valid hostname (ASCII, no scheme/port/path).
func IsHostnameLike(s string) bool {
	// Hostnames must be 1..253 chars.
	if len(s) == 0 || len(s) > 253 {
		return false
	}
	for label := range strings.SplitSeq(s, ".") {
		// Labels must be 1..63 chars.
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		for i := range len(label) {
			ch := label[i]
			// Allow only ASCII letters/digits and hyphen; hyphen can't be first/last.
			if ch == '-' {
				// Allow hyphen only at start/end.
				if i == 0 || i == len(label)-1 {
					return false
				}
				continue
			}
			if !isAlphaNum(ch) {
				return false
			}
		}
	}
	return true
}

// isAlphaNum returns true if the byte is an ASCII letter or digit.
func isAlphaNum(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9')
}
