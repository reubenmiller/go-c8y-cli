package iterator

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// InputCursor holds the current raw input item shared by every
// InputRefIterator of a single command. One input driver advances the cursor
// once per iteration; any number of reference iterators read it. This is what
// lets multiple flags resolve from the same piped item (e.g.
// `--type _.name --name _.name`), which a single per-flag pipe binding cannot.
type InputCursor struct {
	raw []byte
}

// Set updates the current input item.
func (c *InputCursor) Set(raw []byte) { c.raw = raw }

// Raw returns the current input item.
func (c *InputCursor) Raw() []byte { return c.raw }

// InputReference reports whether a flag value references the shared input
// item and returns the gjson path to extract. "-" references the whole item;
// "-.a.b" references the nested path "a.b". These match the existing
// flags.FlagReadFromPipeText ("-") and flags.FlagReadFromPipeJSON ("-.")
// pipe-binding sentinels (kept as literals here to avoid an import cycle).
func InputReference(value string) (path string, ok bool) {
	if value == "-" {
		return "", true
	}
	if strings.HasPrefix(value, "-.") {
		return value[2:], true
	}
	return "", false
}

// InputRefIterator resolves a gjson path against the shared cursor on each
// iteration, applying the flag's format. An empty path yields the whole raw
// item. It is "bound" so the surrounding query/body template re-evaluates it
// per input item, but it never ends iteration itself — the input driver owns
// EOF.
type InputRefIterator struct {
	cursor *InputCursor
	path   string
	format string
}

// NewInputRefIterator returns an iterator bound to cursor that extracts path
// (empty = whole item) and applies format (e.g. "(name eq '%s')") to the
// resolved value.
func NewInputRefIterator(cursor *InputCursor, path, format string) *InputRefIterator {
	if format == "" {
		format = "%s"
	}
	return &InputRefIterator{cursor: cursor, path: path, format: format}
}

// GetNext returns the formatted value extracted from the current input item.
// A missing path yields an empty value (the caller skips it) rather than an
// error, so an optional flag referencing an absent property is simply unset.
func (it *InputRefIterator) GetNext() (line []byte, input interface{}, err error) {
	raw := it.cursor.Raw()
	if it.path == "" {
		if len(raw) == 0 {
			return []byte{}, raw, nil
		}
		return fmt.Appendf(nil, it.format, raw), raw, nil
	}
	v := gjson.GetBytes(raw, it.path)
	if !v.Exists() {
		return []byte{}, raw, nil
	}
	return fmt.Appendf(nil, it.format, v.String()), raw, nil
}

// IsBound always returns true: the value is sourced from the per-item input.
func (it *InputRefIterator) IsBound() bool { return true }
