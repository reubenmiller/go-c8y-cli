package codegen

import (
	"regexp"
	"strings"
)

var dashLetter = regexp.MustCompile(`-(\p{L})`)

// upperFirst capitalizes the first character (invariant culture).
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// lowerFirst lowercases the first character (invariant culture).
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// goCamelCase reproduces the PowerShell expression
//
//	($Name[0].ToString().ToUpperInvariant() + $Name.Substring(1)) -replace '-(\p{L})', { upper }
//
// e.g. "delete-collection" -> "DeleteCollection".
func goCamelCase(name string) string {
	s := upperFirst(name)
	return dashLetter.ReplaceAllStringFunc(s, func(m string) string {
		return strings.ToUpper(m[1:])
	})
}

// goPackageName lowercases a name and replaces dashes with underscores.
func goPackageName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "-", "_")
}

// baseName mimics Split-Path -Leaf for forward slash separated names.
func baseName(name string) string {
	if idx := strings.LastIndexAny(name, "/\\"); idx >= 0 {
		return name[idx+1:]
	}
	return name
}

// matchesTrueYes reproduces the PowerShell expression ($v -match "true|yes"),
// a case-insensitive substring match.
func matchesTrueYes(v string) bool {
	lower := strings.ToLower(v)
	return strings.Contains(lower, "true") || strings.Contains(lower, "yes")
}

// matchesTrue reproduces ($v -match "true").
func matchesTrue(v string) bool {
	return strings.Contains(strings.ToLower(v), "true")
}

// q wraps a value in double quotes without escaping, exactly like the
// PowerShell template interpolation ("`"$value`"") does.
func q(s string) string {
	return `"` + s + `"`
}

// quoteAll renders each value as a double quoted Go string literal token.
func quoteAll(values []string) []string {
	quoted := make([]string, 0, len(values))
	for _, v := range values {
		quoted = append(quoted, `"`+v+`"`)
	}
	return quoted
}
