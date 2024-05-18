package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

func ConvertWildcardToRegex(pattern string) (*regexp.Regexp, error) {
	pattern = strings.ReplaceAll(pattern, "\\", "\\\\")
	pattern = strings.ReplaceAll(pattern, ".", "\\.")

	// convert wildcards to a regex
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	// case-insensitive matching, multi-line matching and . matches \n
	// and whole string matching
	pattern = "(?ims)^" + pattern + "$"

	return regexp.Compile(pattern)
}

func MatchWithWildcards(s, pattern string) (bool, error) {
	r, err := ConvertWildcardToRegex(pattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex patter")
	}

	return r.MatchString(s), nil
}

func MatchWithRegex(s, pattern string) (bool, error) {
	// case-insensitive matching, multi-line matching and . matches \n
	pattern = "(?ims)" + pattern

	r, err := regexp.Compile(pattern)

	if err != nil {
		return false, fmt.Errorf("invalid regex pattern")
	}

	return r.MatchString(s), nil
}
