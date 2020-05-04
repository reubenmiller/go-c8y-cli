package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

func MatchWithWildcards(s, pattern string) (bool, error) {
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, ".", "\\.")

	// convert wildcards to a regex
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	// case insensitive matching and whole string matching
	pattern = "(?i)^" + pattern + "$"

	r, err := regexp.Compile(pattern)

	if err != nil {
		return false, fmt.Errorf("invalid regex patter")
	}

	return r.MatchString(s), nil
}

func MatchWithRegex(s, pattern string) (bool, error) {
	// case-insensitive matching
	pattern = "(?i)" + pattern

	r, err := regexp.Compile(pattern)

	if err != nil {
		return false, fmt.Errorf("invalid regex pattern")
	}

	return r.MatchString(s), nil
}
