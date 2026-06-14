package jsonfilter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/jsondoc"
	outputfilter "github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/filter"
	"github.com/reubenmiller/go-c8y/v2/pkg/c8y/output/shape"
	"github.com/tidwall/gjson"
	"go.uber.org/zap/zapcore"
)

var Logger *logger.Logger

func init() {
	Logger = logger.NewLogger("jsonfilter", logger.Options{
		Level:  zapcore.DebugLevel,
		Color:  true,
		Silent: true,
	})
}

type JSONFilters struct {
	Logger             *logger.Logger
	Filters            []JSONFilter
	Pluck              []string
	Flatten            bool
	AsCSV              bool
	AsTSV              bool
	AsCompletionFormat bool

	// SkipHeaders indicates that the output format does not consume the
	// per-row header callback (e.g. json output), allowing rows that select
	// every property (globstar) to be passed through without being
	// flattened just to resolve header names.
	SkipHeaders bool
}

// KeyGroup records which concrete json keys a single select pattern resolved to
// within one row. Keys is empty when the pattern did not match anything in the row
type KeyGroup = shape.KeyGroup

// HeaderFunc callback which receives the resolved keys of a single row.
// keys is the flat comma separated list (unmatched patterns are included as-is),
// and groups contains the keys grouped by the select pattern which resolved them
type HeaderFunc func(keys string, groups []KeyGroup)

// MergeKeyGroups merges the resolved keys of multiple sampled rows into a single
// ordered column list. Columns are ordered by select pattern, and each pattern
// expands to the union of the keys it resolved to across all rows (deduplicated
// case insensitively). Patterns which did not resolve against any row are
// included as-is so the user can still see which selector returned nothing
func MergeKeyGroups(rows [][]KeyGroup) []string {
	patterns := []string{}
	keysByPattern := map[string][]string{}
	seenKeys := map[string]map[string]bool{}

	for _, row := range rows {
		for _, group := range row {
			if _, ok := keysByPattern[group.Pattern]; !ok {
				patterns = append(patterns, group.Pattern)
				keysByPattern[group.Pattern] = []string{}
				seenKeys[group.Pattern] = map[string]bool{}
			}
			for _, key := range group.Keys {
				keyl := strings.ToLower(key)
				if !seenKeys[group.Pattern][keyl] {
					seenKeys[group.Pattern][keyl] = true
					keysByPattern[group.Pattern] = append(keysByPattern[group.Pattern], key)
				}
			}
		}
	}

	columns := []string{}
	for _, pattern := range patterns {
		if keys := keysByPattern[pattern]; len(keys) > 0 {
			columns = append(columns, keys...)
		} else {
			columns = append(columns, pattern)
		}
	}
	return columns
}

func (f JSONFilters) Apply(jsonValue string, property string, showHeaders bool, setHeaderFunc HeaderFunc) ([]byte, error) {
	return f.filterJSON(jsonValue, property, showHeaders, setHeaderFunc)
}

func splitFilter(s string, sep rune, maxSplit int) []string {
	quoted := false
	openedQuote := ' ' // Use a default character which is not valid quote
	a := strings.FieldsFunc(s, func(r rune) bool {
		// Support both single and double quotes surrounding a string
		if r == '"' || r == '\'' {
			if !quoted {
				openedQuote = r
				quoted = !quoted
			} else if r == openedQuote {
				quoted = !quoted
			}
		}
		return !quoted && r == sep
	})

	// Join rest columns
	if len(a) > maxSplit {
		b := a[0 : maxSplit-1]
		b = append(b, strings.Join(a[maxSplit-1:], string(sep)))
		return b
	}
	return a
}

// AddRawFilters add list of raw filters
func (f *JSONFilters) AddRawFilters(rawFilters []string) error {
	for _, item := range rawFilters {

		property := ""
		operator := ""
		value := ""

		fields := splitFilter(item, ' ', 3)
		switch len(fields) {
		case 0, 1:
		case 2:
			operator = fields[0]
			value = fields[1]
		default: // len > 3
			property = fields[0]
			operator = fields[1]
			value = fields[2]
		}

		if operator == "" {
			operator = "contains"
		}

		if isQuotedString(value) {
			f.Add(property, operator, strings.Trim(value, "\"'"))
			continue
		}

		operatorAliases := map[string]string{
			"has":     "keyIn",
			"hasnot":  "keyNotIn",
			"nothas":  "keyNotIn",
			"missing": "keyNotIn",
		}

		if realName, ok := operatorAliases[operator]; ok {
			operator = realName
		}

		if v, err := strconv.ParseFloat(value, 64); err == nil {
			if strings.Contains(value, ".") {
				// use float
				f.Add(strings.TrimSpace(property), operator, v)
			} else {
				// use int (required by jsonq in some cases, i.e. array length operators like leneq etc.)
				f.Add(strings.TrimSpace(property), operator, int(v))
			}
		} else if v, err := strconv.ParseBool(value); err == nil {
			// Check boolean values
			f.Add(strings.TrimSpace(property), operator, bool(v))
		} else {
			if property == "" {
				// Support keyIn and keyNotIn operators which don't take
				if strings.Contains(value, ".") {
					lastIdx := strings.LastIndex(value, ".")
					property = value[0:lastIdx]
					value = value[lastIdx+1:]
				} else {
					// Default to root element
					property = "."
				}
			}

			f.Add(strings.TrimSpace(property), operator, value)
		}
	}
	return nil
}

func isQuotedString(v string) bool {
	return (strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"")) || (strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'"))
}

func (f *JSONFilters) Add(property, operation string, value interface{}) {
	f.Filters = append(f.Filters, JSONFilter{
		Property:  property,
		Operation: operation,
		Value:     value,
	})
}

// NewJSONFilters create a json filter
func NewJSONFilters(l *logger.Logger) *JSONFilters {
	return &JSONFilters{
		Logger:  l,
		Filters: make([]JSONFilter, 0),
	}
}

type JSONFilter struct {
	Property  string
	Operation string
	Value     interface{}
}

// filterJSON applies filtering and property selection using the streaming
// engine from go-c8y/v2 (compiled predicates, single pass over the items).
func (f JSONFilters) filterJSON(jsonValue string, property string, showHeaders bool, setHeaderFunc HeaderFunc) ([]byte, error) {
	pred, err := f.compilePredicate()
	if err != nil {
		return nil, err
	}
	return f.filterJSONStreaming(pred, jsonValue, property, showHeaders, setHeaderFunc)
}

// compilePredicate compiles the filters to a single predicate using the
// go-c8y/v2 filter engine. All pattern/regex/date compilation happens once
// here rather than per row.
func (f JSONFilters) compilePredicate() (outputfilter.Predicate, error) {
	preds := make([]outputfilter.Predicate, 0, len(f.Filters))
	for _, q := range f.Filters {
		p, err := outputfilter.Condition(q.Property, q.Operation, q.Value)
		if err != nil {
			return nil, fmt.Errorf("filter error. filter=%s %s %v: %w", q.Property, q.Operation, q.Value, err)
		}
		preds = append(preds, p)
	}
	return outputfilter.And(preds...), nil
}

// filterJSONStreaming filters and shapes the items in a single pass using
// compiled predicates, avoiding a decode/re-marshal round trip of the whole
// payload.
func (f JSONFilters) filterJSONStreaming(pred outputfilter.Predicate, jsonValue string, property string, showHeaders bool, setHeaderFunc HeaderFunc) ([]byte, error) {
	v := gjson.Parse(jsonValue)
	if property != "" {
		v = v.Get(property)
	}

	isObject := v.IsObject()
	if !isObject && !v.IsArray() {
		return []byte(v.Str), nil
	}

	usePluck := len(f.Pluck) > 0 || f.Flatten
	// When every property is selected and the output format does not use
	// per-row headers, rows can be passed through untouched instead of being
	// flattened (the flatten exists only to resolve header names).
	rawPassthrough := f.SkipHeaders && !f.AsCSV && !f.AsTSV && !f.AsCompletionFormat && !f.Flatten &&
		len(f.Pluck) == 1 && strings.TrimSpace(f.Pluck[0]) == "**"

	// Compile the property selection once; per item the cost is a single
	// flatten+match pass instead of re-compiling the glob patterns per row
	var selector *shape.Selector
	if usePluck && !rawPassthrough {
		selector = shape.NewSelector(f.Pluck...)
	}

	outputValues := make([]string, 0)
	matchedRaw := make([]string, 0)
	headerDone := !showHeaders

	handleItem := func(item gjson.Result) {
		if !pred(jsondoc.New([]byte(item.Raw))) {
			return
		}
		if !usePluck {
			matchedRaw = append(matchedRaw, item.Raw)
			return
		}
		if rawPassthrough {
			outputValues = append(outputValues, item.Raw)
			return
		}
		if item.IsObject() {
			if !headerDone {
				headerDone = true
				outputValues = append(outputValues, expandHeaderProperties(&item, f.Pluck))
			}
			if line, keys, keyGroups := f.applySelection(selector, &item); line != "" {
				outputValues = append(outputValues, line)
				setHeaderFunc(strings.Join(keys, ","), keyGroups)
			}
		} else {
			outputValues = append(outputValues, item.Raw)
		}
	}

	if isObject {
		handleItem(v)
	} else {
		v.ForEach(func(_, item gjson.Result) bool {
			handleItem(item)
			return true
		})
	}

	if usePluck {
		return []byte(strings.Join(outputValues, "\n")), nil
	}

	if isObject {
		if len(matchedRaw) > 0 {
			return []byte(matchedRaw[0]), nil
		}
		return []byte(""), nil
	}

	// JSON array output. The legacy engine's encoder emits a trailing
	// newline which downstream empty-result detection relies on.
	var b bytes.Buffer
	b.WriteByte('[')
	for i, raw := range matchedRaw {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(raw)
	}
	b.WriteString("]\n")
	return b.Bytes(), nil
}

// maxUnflattenKeyCount protects against rebuilding json with an excessive
// number of keys; such rows are returned unshaped instead
const maxUnflattenKeyCount = 10000

// applySelection renders a single row using the compiled selector,
// honouring the configured output mode (json, flat json, csv, tsv or
// completion format). It returns an empty line when the row can not be
// processed (matching the legacy pluck behavior)
func (f JSONFilters) applySelection(selector *shape.Selector, item *gjson.Result) (string, []string, []KeyGroup) {
	sel, err := selector.Apply([]byte(item.Raw))
	if err != nil {
		return "", nil, nil
	}

	line := ""
	switch {
	case f.AsCSV:
		line = sel.CSV(",")
	case f.AsTSV:
		line = sel.CSV("\t")
	case f.AsCompletionFormat:
		line = convertSelectionToLine(sel)
	case f.Flatten:
		if v, err := sel.FlatJSON(); err == nil {
			line = string(v)
		} else {
			Logger.Warningf("failed to marshal value. err=%s", err)
		}
	default:
		switch {
		case selector.SelectsEverything():
			line = item.Raw
		case sel.Size() > maxUnflattenKeyCount:
			if f.Logger != nil {
				itemID := ""
				if v := item.Get("id"); v.Exists() {
					itemID = v.Str
				}
				f.Logger.Warnf("Detected json with a large number of keys, returning all data by default. Use jq for further filtering. total_keys=%d, id=%s", sel.Size(), itemID)
			}
			line = item.Raw
		default:
			if v, err := sel.JSON(); err == nil {
				line = string(v)
			} else {
				Logger.Warningf("failed to marshal value. err=%s", err)
			}
		}
	}
	return line, sel.Keys(), sel.Groups()
}

// convertSelectionToLine renders a selection in the completion format:
// "{value}\t{key1}: {value1} | {key2}: {value2}"
func convertSelectionToLine(sel *shape.Selection) string {
	buf := bytes.Buffer{}
	for i, key := range sel.Keys() {
		if i != 0 {
			// handle for empty non-existent values by leaving it blank
			if i == 1 {
				buf.WriteString("\t")
			} else {
				buf.WriteString(" | ")
			}
		}
		if value, ok := sel.Value(key); ok {
			marshalledValue, err := json.Marshal(value)
			if err != nil {
				Logger.Warningf("failed to marshal value. value=%v, err=%s", value, err)
				continue
			}
			if i != 0 {
				buf.WriteString(key)
				buf.WriteString(": ")
			}
			if !bytes.Contains(marshalledValue, []byte(",")) {
				buf.Write(bytes.Trim(marshalledValue, "\""))
			} else {
				buf.Write(marshalledValue)
			}
		}
	}
	return buf.String()
}

func expandHeaderProperties(item *gjson.Result, properties []string) string {
	headers := []string{}

	if item.IsArray() {
		items := item.Array()
		if len(items) > 0 {
			item = &items[0]
		}
	}

	for _, key := range properties {
		name, _, err := resolveKeyName(item, key)
		if err != nil {
			headers = append(headers, key)
		} else {
			headers = append(headers, name)
		}
	}
	return strings.Join(headers, ",")
}

func resolveKeyName(item *gjson.Result, key string) (name string, value interface{}, err error) {
	if value := item.Get(key); value.Exists() {
		// Here: How to get t
		//
		tokenEnd := strings.LastIndex(item.Raw[:value.Index], "\"")
		if tokenEnd == -1 {
			return key, nil, nil
		}
		tokenStart := strings.LastIndex(item.Raw[:tokenEnd], "\"")

		if tokenStart == -1 {
			return key, nil, nil
		}
		return item.Raw[tokenStart+1 : tokenEnd], value.Value(), nil
	}
	return key, nil, nil
}
