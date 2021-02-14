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
	Pluck     []string
	AsCSV     bool
}

func (f JSONFilters) Apply(jsonValue string, property string, showHeaders bool) []byte {
	return filterJSON(jsonValue, property, f, f.Selectors, f.Pluck, f.AsCSV, showHeaders)
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

func filterJSON(jsonValue string, property string, filters JSONFilters, selectors []string, pluckValues []string, asCSV bool, showHeaders bool) []byte {
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
	jq.Macro("-like", matchWithWildcards)
	jq.Macro("match", matchWithRegex)
	jq.Macro("-match", matchWithRegex)

	for _, query := range filters.Filters {
		Logger.Debugf("filtering data: %s %s %s", query.Property, query.Operation, query.Value)
		jq.Where(query.Property, query.Operation, query.Value)
	}

	if len(selectors) > 0 {
		jq.Select(selectors...)
	}

	// format values (using gjson)
	if len(pluckValues) > 0 {
		var bsub bytes.Buffer
		jq.Writer(&bsub)
		formattedJSON := gjson.ParseBytes(bsub.Bytes())

		if formattedJSON.IsArray() {
			outputValues := make([]string, 0)

			if showHeaders {
				outputValues = append(outputValues, expandHeaderProperties(&formattedJSON, pluckValues))
			}

			for _, myval := range formattedJSON.Array() {
				Console.SetHeaderFromInput(myval.String())
				if line := pluckJsonValues(&myval, pluckValues, asCSV); line != "" {
					outputValues = append(outputValues, line)
				}
			}
			return []byte(strings.Join(outputValues, "\n"))
		}

		Console.SetHeaderFromInput(formattedJSON.String())

		if line := pluckJsonValues(&formattedJSON, pluckValues, asCSV); line != "" {
			headers := ""
			if showHeaders {
				headers = expandHeaderProperties(&formattedJSON, pluckValues)
			}
			if headers != "" {
				headers += "\n"
			}
			return []byte(headers + line)
		}

		Logger.Debugf("ERROR: gjson path does not exist. %v", pluckValues)
		return []byte("")
	}

	jq.Writer(&b)

	// Convert back to an object if it
	if convertBackFromArray {
		return removeJSONArrayValues(b.Bytes())
	}
	return b.Bytes()
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

// func main() {
// 	item := gjson.Parse(`{
// 		"details": {
// 			"someReallyLongName": "RUNNING"
// 		}
// 	}
// 	`)
// 	realKeyName, value, err := resolveKeyName(item, "det*.some*")
// 	if err != nil {
// 		fmt.Printf("%s: %s", realKeyName, value)
// 	}
// }

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
	// return "", nil, fmt.Errorf("not found. key=%s", key)
}

func pluckJsonValues(item *gjson.Result, properties []string, asCSV bool) string {
	if item == nil {
		return ""
	}
	values := []string{}

	// json line output
	if !asCSV {
		for _, prop := range properties {
			if !strings.ContainsAny(prop[:1], "{[") {
				prop = "{" + strings.Trim(prop, "{}") + "}"
			}
			// complexSelector := "{" + strings.Trim(prop, "{}") + "}"
			if v := item.Get(prop); v.Exists() {
				values = append(values, v.String())
			}
		}
		// allow output fan out
		return strings.Join(values, "\n")
	}

	columns := []string{}
	for _, prop := range properties {
		columns = append(columns, strings.Split(prop, ",")...)
	}
	// Csv output
	for _, pluck := range columns {
		// strip alias in csv, the columns are handled elsewhere name:value
		if index := strings.Index(pluck, ":"); index != -1 {
			pluck = pluck[index+1:]
		}
		if v := item.Get(pluck); v.Exists() {
			values = append(values, v.String())
		} else {
			// add empty value
			values = append(values, "")
		}
	}
	return strings.Join(values, ",")
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

func getFilterFlag(cmd *cobra.Command, flagName string) *JSONFilters {
	filters := newJSONFilters()

	enableCSV := false
	if cmd.Flags().Changed("csv") {
		if value, err := cmd.Flags().GetBool("csv"); err == nil {
			enableCSV = value
		}
	}

	if cmd.Flags().Changed("select") {
		if properties, err := cmd.Flags().GetStringArray("select"); err == nil {
			formattedProperties := []string{}

			for _, prop := range properties {
				formattedProperties = append(formattedProperties, prop)
			}
			filters.AsCSV = enableCSV
			filters.Pluck = formattedProperties
		}
	}

	if cmd.Flags().Changed(flagName) {
		if rawFilters, err := cmd.Flags().GetStringSlice(flagName); err == nil {
			for _, item := range rawFilters {
				sepPattern := regexp.MustCompile("(\\s+[\\-]?(like|match|eq|neq|lt|lte|gt|gte|notIn|in|startsWith|endsWth|contains|len[n]?eq|lengt[e]?|lenlt[e]?)\\s+|(!?=|[<>]=?))")

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
