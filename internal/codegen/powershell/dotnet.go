package powershell

import (
	"regexp"
	"strings"
)

// dotnetReplace ports the PowerShell -replace operator: every match of re in
// input is replaced with replacement, expanding the .NET substitution
// patterns in the replacement string. This matters because the generator
// templates are substituted with user supplied text (example commands) that
// can contain sequences like $_ — which .NET expands to the entire input
// string — and the committed output depends on those semantics.
//
// The generator patterns define no capturing groups, so group references
// other than $0 stay literal (like .NET treats references to undefined
// groups).
func dotnetReplace(re *regexp.Regexp, input, replacement string) string {
	matches := re.FindAllStringIndex(input, -1)
	if matches == nil {
		return input
	}
	b := &strings.Builder{}
	last := 0
	for _, m := range matches {
		b.WriteString(input[last:m[0]])
		b.WriteString(expandSubstitution(replacement, input, m[0], m[1]))
		last = m[1]
	}
	b.WriteString(input[last:])
	return b.String()
}

// expandSubstitution renders a .NET substitution string for one match.
func expandSubstitution(replacement, input string, start, end int) string {
	b := &strings.Builder{}
	for i := 0; i < len(replacement); i++ {
		c := replacement[i]
		if c != '$' || i+1 >= len(replacement) {
			b.WriteByte(c)
			continue
		}
		switch next := replacement[i+1]; {
		case next == '$':
			b.WriteByte('$')
			i++
		case next == '&':
			b.WriteString(input[start:end])
			i++
		case next == '`':
			b.WriteString(input[:start])
			i++
		case next == '\'':
			b.WriteString(input[end:])
			i++
		case next == '_':
			b.WriteString(input)
			i++
		case next >= '0' && next <= '9':
			j := i + 1
			for j < len(replacement) && replacement[j] >= '0' && replacement[j] <= '9' {
				j++
			}
			if number := replacement[i+1 : j]; strings.TrimLeft(number, "0") == "" {
				// $0 is the whole match; other groups do not exist and stay literal
				b.WriteString(input[start:end])
				i = j - 1
			} else {
				b.WriteByte(c)
			}
		case next == '{':
			closing := strings.IndexByte(replacement[i:], '}')
			if closing < 0 {
				b.WriteByte(c)
				continue
			}
			if name := replacement[i+2 : i+closing]; strings.TrimLeft(name, "0") == "" && name != "" {
				b.WriteString(input[start:end])
				i += closing
			} else {
				b.WriteByte(c)
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
