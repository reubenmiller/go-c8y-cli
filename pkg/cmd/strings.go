package cmd

import (
	"strings"
)

// RemoveEmptyStrings returns a new array where the strings are not empty (after trimming space)
func RemoveEmptyStrings(array []string) []string {
	out := []string{}

	for _, val := range array {
		val = strings.TrimSpace(val)
		if val != "" {
			out = append(out, val)
		}
	}
	return out
}

func SplitString(value string, sep string) []string {
	return RemoveEmptyStrings(strings.Split(value, sep))
}
