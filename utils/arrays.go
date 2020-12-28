package utils

import "strings"

// ArrayContainsStringIgnoreCase check if a string array contains target string, ignore case
func ArrayContainsStringIgnoreCase(src []string, target string) bool {
	for _, s := range src {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}
