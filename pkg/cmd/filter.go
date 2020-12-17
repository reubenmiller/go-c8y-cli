package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq"
	"github.com/tidwall/gjson"
)

type JSONFilters struct {
	Filters   []JSONFilter
	Selectors []string
	Pluck     string
}

func (f JSONFilters) Apply(jsonValue string, property string) []byte {
	return filterJSON(jsonValue, property, f, f.Selectors, f.Pluck)
}

func (f *JSONFilters) AddSelectors(props ...string) {
	f.Selectors = append(f.Selectors, props...)
}

func (f *JSONFilters) Add(property, operation, value string) {
	f.Filters = append(f.Filters, JSONFilter{
		Property:  property,
		Operation: operation,
		Value:     value,
	})
}

func newJSONFilters() *JSONFilters {
	return &JSONFilters{
		Filters:   make([]JSONFilter, 0),
		Selectors: make([]string, 0),
	}
}

type JSONFilter struct {
	Property  string
	Operation string
	Value     string
}

func isJSONArrayString(jsonValue string) bool {
	trimmed := strings.TrimSpace(jsonValue)
	Logger.Debugf("checking string: %s, first=%v, last=%v", trimmed, trimmed[0], trimmed[len(trimmed)-1])
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

func isJSONArrayBytes(jsonValue []byte) bool {
	trimmed := bytes.TrimSpace(jsonValue)
	return bytes.HasPrefix(trimmed, []byte("[")) && bytes.HasSuffix(trimmed, []byte("]"))
}

func removeJSONArrayValues(jsonValue []byte) []byte {
	v := gjson.ParseBytes(jsonValue)
	if !v.IsArray() {
		return jsonValue
	}
	if len(v.Array()) > 0 {
		return []byte(v.Array()[0].String())
	}
	return []byte("")
}

func filterJSON(jsonValue string, property string, filters JSONFilters, selectors []string, pluck string) []byte {
	var b bytes.Buffer

	var jq *gojsonq.JSONQ
	convertBackFromArray := false

	v := gjson.Parse(jsonValue)

	if property != "" {
		v = v.Get(property)
	}

	if v.IsObject() {
		Logger.Info("Converting json object to array")
		jq = gojsonq.New().FromString("[" + v.String() + "]")
		convertBackFromArray = true
	} else {
		jq = gojsonq.New().FromString(v.String())
	}

	// Add custom filters
	jq.Macro("like", matchWithWildcards)
	jq.Macro("match", matchWithRegex)

	for _, query := range filters.Filters {
		Logger.Debugf("filtering data: %s %s %s", query.Property, query.Operation, query.Value)
		jq.Where(query.Property, query.Operation, query.Value)
	}

	if len(selectors) > 0 {
		jq.Select(selectors...)
	}

	// format values (using gjson)
	if pluck != "" {
		var bsub bytes.Buffer
		jq.Writer(&bsub)
		formattedJSON := gjson.ParseBytes(bsub.Bytes())

		if formattedJSON.IsArray() {
			outputValues := make([]string, 0)
			for _, myval := range formattedJSON.Array() {
				if v := myval.Get(pluck); v.Exists() {
					outputValues = append(outputValues, v.String())
				}
			}
			return []byte(strings.Join(outputValues, "\n"))
		}
		if v := formattedJSON.Get(pluck); v.Exists() {
			return []byte(strings.TrimRight(v.String(), "\n"))
		}
		Logger.Debugf("ERROR: gjson path does not exist. %s", pluck)
		return []byte("")
	}

	jq.Writer(&b)

	// Convert back to an object if it
	if convertBackFromArray {
		return removeJSONArrayValues(b.Bytes())
	}
	return b.Bytes()
}

func matchWithWildcards(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	return matcher.MatchWithWildcards(xs, pattern)
}

func matchWithRegex(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	return matcher.MatchWithRegex(xs, pattern)
}

func addFilterFlag(cmd *cobra.Command, name string) {
	if name == "" {
		name = "filter"
	}
	cmd.Flags().StringSlice(name, nil, "filter")
	cmd.Flags().StringSlice("select", nil, "select")
	cmd.Flags().String("format", "", "format")
}

func getFilterFlag(cmd *cobra.Command, flagName string) *JSONFilters {
	filters := newJSONFilters()

	if cmd.Flags().Changed("select") {
		if props, err := cmd.Flags().GetStringSlice("select"); err == nil {
			filters.AddSelectors(props...)
		}
	}

	if cmd.Flags().Changed("format") {
		if prop, err := cmd.Flags().GetString("format"); err == nil {
			filters.Pluck = prop
		}
	}

	if cmd.Flags().Changed(flagName) {
		if rawFilters, err := cmd.Flags().GetStringSlice(flagName); err == nil {
			for _, item := range rawFilters {
				sepPattern := regexp.MustCompile("(\\s+(like|match|eq|neq|lt|lte|gt|gte|notIn|in|startsWith|endsWth|contains|len[n]?eq|lengt[e]?|lenlt[e]?)\\s+|(!?=|[<>]=?))")

				parts := sepPattern.Split(item, 2)

				if len(parts) != 2 {
					continue
				}
				operator := sepPattern.FindString(item)

				if operator == "" {
					operator = "contains"
				}
				operator = strings.TrimSpace(operator)
				operator = strings.ReplaceAll(operator, " ", "")
				filters.Add(strings.TrimSpace(parts[0]), operator, strings.TrimSpace(parts[1]))
			}
		}
	}

	return filters
}
