package libcrawl

import (
	"strings"
)

func smallestSubstrRight(s string, delimiter string) string {
	li := strings.LastIndex(s, delimiter)
	if li+1 < len(s) {
		return s[li+1:]
	}
	return ""
}

func isMember(list []string, s string) int {
	for i, v := range list {
		if v == s {
			return i
		}
	}
	return -1
}
