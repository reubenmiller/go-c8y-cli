package jsonfilter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/facette/natsort"
	glob "github.com/obeattie/ohmyglob"
	"github.com/reubenmiller/go-c8y-cli/pkg/flatten"
	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/thedevsaddam/gojsonq"
	"github.com/tidwall/gjson"
)

var Logger logger.Logger

type JSONFilters struct {
	Filters   []JSONFilter
	Selectors []string
	Pluck     []string
	Flatten   bool
	AsCSV     bool
}

func (f JSONFilters) Apply(jsonValue string, property string, showHeaders bool, setHeaderFunc func(string)) []byte {
	return f.filterJSON(jsonValue, property, showHeaders, setHeaderFunc)
}

func (f *JSONFilters) AddSelectors(props ...string) {
	f.Selectors = append(f.Selectors, props...)
}

// AddRawFilters add list of raw filters
func (f *JSONFilters) AddRawFilters(rawFilters []string) error {
	for _, item := range rawFilters {
		sepPattern := regexp.MustCompile(`(\s+[\-]?(like|match|eq|neq|lt|lte|gt|gte|notIn|in|startsWith|endsWth|contains|len[n]?eq|lengt[e]?|lenlt[e]?)\s+|(!?=|[<>]=?))`)

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
		f.Add(strings.TrimSpace(parts[0]), operator, strings.TrimSpace(parts[1]))
	}
	return nil
}

func (f *JSONFilters) Add(property, operation, value string) {
	f.Filters = append(f.Filters, JSONFilter{
		Property:  property,
		Operation: operation,
		Value:     value,
	})
}

// FilterPropertyByWildcard filtery a json string by using globstar (wildcards) on the nested json paths
func FilterPropertyByWildcard(jsonValue string, prefix string, patterns []string, setAlias bool) (map[string]interface{}, []string, error) {
	rawMap := make(map[string]interface{})
	err := c8y.DecodeJSONBytes([]byte(jsonValue), &rawMap)
	if err != nil {
		return nil, nil, err
	}
	flatMap, err := flatten.Flatten(rawMap, prefix, flatten.DotStyle)

	if err != nil {
		return nil, nil, err
	}
	compiledPatterns := []glob.Glob{}
	aliases := []string{}
	filteredMap := make(map[string]interface{})

	for _, p := range patterns {
		// resolve path using wildcards
		// strip alias reference
		alias := ""
		if idx := strings.Index(p, ":"); idx > -1 {
			alias = p[0:idx]
			p = p[idx+1:]
		}
		p = strings.ToLower(p)

		if p != "" {
			if cp, err := glob.Compile(p, &glob.Options{
				Separator:    '.',
				MatchAtStart: true,
				MatchAtEnd:   true,
			}); err == nil {
				compiledPatterns = append(compiledPatterns, cp)
				aliases = append(aliases, alias)
			}
		}
	}

	resolvedProperties, _ := filterFlatMap(flatMap, filteredMap, compiledPatterns, aliases)
	return filteredMap, resolvedProperties, err
}

func filterFlatMap(src map[string]interface{}, dst map[string]interface{}, patterns []glob.Glob, aliases []string) ([]string, error) {
	sortedKeys := []string{}

	// sort source map keys, so matching is stable when using alias
	// where one pattern can match multiple values (otherwise it will be random which property is selected)
	sourceKeys := make([]string, len(src))
	i := 0
	for key := range src {
		sourceKeys[i] = key
		i++
	}
	// Use natural sorting to sory array in an user friendly way
	// i.e. 1, 10, 2 => 1, 2, 10
	natsort.Sort(sourceKeys)

	for i, pattern := range patterns {
		found := false
		for _, key := range sourceKeys {
			value := src[key]
			keyl := strings.ToLower(key)
			if strings.HasPrefix(keyl, pattern.String()+".") || pattern.MatchString(keyl) {
				if aliases[i] != "" {
					paths := strings.Split(pattern.String(), ".")
					if strings.Contains(pattern.String(), "*") {
						commonpath := bytes.Buffer{}
						hasAlias := false

						if strings.HasPrefix(pattern.String(), "*") {
							key = aliases[i] + "." + key
							hasAlias = true
						} else if strings.HasSuffix(pattern.String(), "*") {
							keyPaths := strings.Split(key, ".")
							for idxPart, part := range paths {
								if strings.Contains(part, "**") || part == "*" {
									break
								}
								// get the real key path rather than the wildcard
								if strings.Contains(part, "*") && idxPart < len(keyPaths) {
									part = keyPaths[idxPart]
									commonpath.WriteString("." + part)
									break
								}
								commonpath.WriteString("." + part)
							}
							commonprefix := strings.TrimLeft(commonpath.String(), ".")
							if strings.HasPrefix(keyl, strings.ToLower(commonprefix)) {
								key = aliases[i] + key[len(commonprefix):]
								hasAlias = true
							}
						}

						if !hasAlias {
							key = aliases[i]
						}
					} else {
						key = aliases[i]
					}
				}
				dst[key] = value
				sortedKeys = append(sortedKeys, key)
				found = true
			}
		}
		if !found {
			// store non-matching patterns for csv generation
			sortedKeys = append(sortedKeys, pattern.String())
		}
	}
	return sortedKeys, nil
}

// NewJSONFilters create a json filter
func NewJSONFilters() *JSONFilters {
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

func (f JSONFilters) filterJSON(jsonValue string, property string, showHeaders bool, setHeaderFunc func(string)) []byte {
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

	for _, query := range f.Filters {
		Logger.Debugf("filtering data: %s %s %s", query.Property, query.Operation, query.Value)
		jq.Where(query.Property, query.Operation, query.Value)
	}

	if len(f.Selectors) > 0 {
		jq.Select(f.Selectors...)
	}

	// format values (using gjson)
	if len(f.Pluck) > 0 || f.Flatten {
		var bsub bytes.Buffer
		jq.Writer(&bsub)
		formattedJSON := gjson.ParseBytes(bsub.Bytes())

		if formattedJSON.IsArray() {
			outputValues := make([]string, 0)

			if showHeaders {
				outputValues = append(outputValues, expandHeaderProperties(&formattedJSON, f.Pluck))
			}

			for _, myval := range formattedJSON.Array() {

				if line, keys := pluckJsonValues(&myval, f.Pluck, f.Flatten, f.AsCSV); line != "" {
					outputValues = append(outputValues, line)
					setHeaderFunc(strings.Join(keys, ","))
				}
			}
			return []byte(strings.Join(outputValues, "\n"))
		}

		if line, keys := pluckJsonValues(&formattedJSON, f.Pluck, f.Flatten, f.AsCSV); line != "" {
			setHeaderFunc(strings.Join(keys, ","))
			return []byte(line)
		}

		Logger.Debugf("ERROR: gjson path does not exist. %v", f.Pluck)
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

func pluckJsonValues(item *gjson.Result, properties []string, flat bool, asCSV bool) (string, []string) {
	if item == nil {
		return "", nil
	}

	if len(properties) == 0 {
		properties = append(properties, "**")
	}

	// flatten json
	pathPatterns := make([]string, 0)
	for _, key := range properties {
		pathPatterns = append(pathPatterns, strings.Split(key, ",")...)
	}

	useAliases := false
	flatMap, flatKeys, err := FilterPropertyByWildcard(item.Raw, "", pathPatterns, useAliases)
	if err != nil {
		return "", nil
	}

	output := bytes.Buffer{}

	// json output
	var v []byte
	if flat || asCSV {
		if asCSV {
			return convertToCSV(flatMap, flatKeys), flatKeys
		}
		v, err = json.Marshal(flatMap)
	} else {
		// unflatten
		v, err = flatten.Unflatten(flatMap)
	}
	if err != nil {
		Logger.Warningf("failed to marshal value. err=%s", err)
	} else {
		if v != nil {
			output.Write(v)
		}
	}

	if err != nil {
		return "", nil
	}

	return output.String(), flatKeys
}

func convertToCSV(flatMap map[string]interface{}, keys []string) string {
	buf := bytes.Buffer{}
	for i, key := range keys {
		if i != 0 {
			// handle for empty non-existant values by leaving it blank
			buf.WriteByte(',')
		}
		if value, ok := flatMap[key]; ok {
			if marshalledValue, err := json.Marshal(value); err != nil {
				Logger.Warningf("failed to marshal value. value=%v, err=%s", value, err)
			} else {
				if !bytes.Contains(marshalledValue, []byte(",")) {
					buf.Write(bytes.Trim(marshalledValue, "\""))
				} else {
					buf.Write(marshalledValue)
				}
			}
		}
	}
	return buf.String()
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