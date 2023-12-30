package jsonfilter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/itchyny/gojq"
	glob "github.com/obeattie/ohmyglob"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/flatten"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/matcher"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/sortorder"
	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"github.com/reubenmiller/gojsonq/v2"
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
	Logger               *logger.Logger
	Filters              []JSONFilter
	Pluck                []string
	PostJQ               string
	Flatten              bool
	AsCSV                bool
	AsTSV                bool
	AsCompletionFormat   bool
	UseOldImplementation bool
}

func (f JSONFilters) Apply(jsonValue string, property string, showHeaders bool, setHeaderFunc func(string)) ([]byte, error) {
	if f.UseOldImplementation {
		return f.filterJSON(jsonValue, property, showHeaders, setHeaderFunc)
	}
	return f.ApplyToData(jsonValue, property, showHeaders, setHeaderFunc)
}

func (f JSONFilters) ApplyToData(data any, property string, showHeaders bool, setHeaderFunc func(string)) ([]byte, error) {
	switch v := data.(type) {
	case []byte:
		var d any
		if err := c8y.DecodeJSONBytes(v, &d); err != nil {
			return nil, err
		}
		return f.filterJSONUsingJQ(d, property, showHeaders, setHeaderFunc)
	case string:
		var d any
		if err := c8y.DecodeJSONBytes([]byte(v), &d); err != nil {
			return nil, err
		}
		return f.filterJSONUsingJQ(d, property, showHeaders, setHeaderFunc)

	}
	return f.filterJSONUsingJQ(data, property, showHeaders, setHeaderFunc)
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

// FilterPropertyByWildcard filtery a json string by using globstar (wildcards) on the nested json paths
func FilterPropertyByWildcard(jsonValue string, prefix string, patterns []string, setAlias bool) (map[string]interface{}, []string, error) {
	rawMap := make(map[string]interface{})
	err := c8y.DecodeJSONBytes([]byte(jsonValue), &rawMap)
	if err != nil {
		return nil, nil, err
	}
	Logger.Debugf("flattening json")
	flatMap, err := flatten.Flatten(rawMap, prefix, flatten.DotStyle)
	Logger.Debugf("finished flattening json")

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

	Logger.Debugf("running filterFlatMap")
	resolvedProperties, _ := filterFlatMap(flatMap, filteredMap, compiledPatterns, aliases)
	Logger.Debugf("finished filterFlatMap")
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
	// Use natural sorting to sort array in an user friendly way
	// i.e. 1, 10, 2 => 1, 2, 10
	// Skip sorting for very large key sets
	if len(sourceKeys) <= 2000000 {
		sort.Sort(sortorder.Natural(sourceKeys))
	}

	for i, pattern := range patterns {
		found := false
		Logger.Debugf("filtering keys by pattern: total=%d, pattern=%s", len(sourceKeys), pattern.String())
		for _, key := range sourceKeys {
			value := src[key]

			// noramlize key, and strip the key identifier
			keyl := strings.ReplaceAll(strings.ToLower(key), flatten.KeyPrefix, "")

			if strings.HasPrefix(keyl, pattern.String()+".") || (pattern.MatchString(keyl) && !pattern.IsNegative()) {
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
		if !found && !pattern.IsNegative() {
			// store non-matching patterns for csv generation
			sortedKeys = append(sortedKeys, pattern.String())
		}
	}

	// filter for negated keys
	sortedMatchingKeys := make([]string, 0)
	for _, key := range sortedKeys {
		keyl := strings.ToLower(key)
		match := true
		for _, pattern := range patterns {
			if pattern.IsNegative() {
				if pattern.MatchString(keyl) {
					match = false
					delete(dst, key)
				}
			}
		}
		if match {
			sortedMatchingKeys = append(sortedMatchingKeys, key)
		}
	}

	return sortedMatchingKeys, nil
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

func isJQJSONQNodeError(err error) bool {
	return strings.Contains(err.Error(), "invalid node name")
}

func formatErrors(errs []error) error {
	filteredErrs := make([]error, 0)

	for _, err := range errs {
		// ignore invalid node name errors, as the property does not exist
		if !isJQJSONQNodeError(err) {
			filteredErrs = append(filteredErrs, err)
		}
	}
	if len(filteredErrs) > 0 {
		return fmt.Errorf("filter error. %s", filteredErrs[0])
	}

	return nil
}

func (f JSONFilters) HasPostJQFilter() bool {
	return f.PostJQ != ""
}

func toPathCheckJQ(property string, value string, negate bool) string {
	parts := strings.Split(property, ".")
	paths := make([]string, 0)
	paths = append(paths, parts...)
	paths = append(paths, value)

	sb := strings.Builder{}
	lines := 0
	for _, v := range paths {
		if v != "" {
			if lines > 0 {
				sb.WriteString(`,`)
			}
			sb.WriteString(`"`)
			sb.WriteString(v)
			sb.WriteString(`"`)
			lines = 1
		}
	}

	if negate {
		return fmt.Sprintf("select(IN(paths; [%v]) | not)", sb.String())
	} else {
		return fmt.Sprintf("select(IN(paths; [%v]))", sb.String())
	}
}

func (f JSONFilters) GetJQQuery(property string) string {
	queryParts := []string{}

	// Pre-process
	if property != "" {
		queryParts = append(queryParts, property)
	}

	// Filter
	for _, ff := range f.Filters {
		switch ff.Operation {
		case "like", "-like":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | test(\"%s\"))", ff.Property, strings.ReplaceAll(fmt.Sprintf("%v", ff.Value), "*", ".*")))
		case "notlike", "-notlike":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | test(\"%s\") | not)", ff.Property, strings.ReplaceAll(fmt.Sprintf("%v", ff.Value), "*", ".*")))
		case "match", "-match":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | test(\"%v\"))", ff.Property, ff.Value))
		case "notmatch", "-notmatch":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | test(\"%v\") | not)", ff.Property, ff.Value))
		case "eq", "-eq", "==", "=":
			switch v := ff.Value.(type) {
			case string:
				queryParts = append(queryParts, fmt.Sprintf("select(.%s == \"%v\")", ff.Property, v))
			default:
				queryParts = append(queryParts, fmt.Sprintf("select(.%s == %v)", ff.Property, v))
			}
		case "neq", "-neq", "!=", "<>":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s != \"%v\")", ff.Property, ff.Value))
		case "leneq", "-leneq":
			queryParts = append(queryParts, fmt.Sprintf("select((.%s | length) == %v)", ff.Property, ff.Value))
		case "lengt", "-lengt":
			queryParts = append(queryParts, fmt.Sprintf("select((.%s | length) > %v)", ff.Property, ff.Value))
		case "lengte", "-lengte":
			queryParts = append(queryParts, fmt.Sprintf("select((.%s | length) >= %v)", ff.Property, ff.Value))
		case "lenlt", "-lenlt":
			queryParts = append(queryParts, fmt.Sprintf("select((.%s | length) < %v)", ff.Property, ff.Value))
		case "lenlte", "-lenlte":
			queryParts = append(queryParts, fmt.Sprintf("select((.%s | length) <= %v)", ff.Property, ff.Value))
		case "has", "-has", "keyIn":
			queryParts = append(queryParts, toPathCheckJQ(ff.Property, fmt.Sprintf("%v", ff.Value), false))
		case "hasnot", "-hasnot", "keyNotIn":
			queryParts = append(queryParts, toPathCheckJQ(ff.Property, fmt.Sprintf("%v", ff.Value), true))
			// queryParts = append(queryParts, fmt.Sprintf("select(has(\"%v\") | not)", ff.Value))
		case "gt", "-gt", ">":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s > %v)", ff.Property, ff.Value))
		case "gte", "-gte", ">=":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s >= %v)", ff.Property, ff.Value))
		case "lt", "-lt", "<":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s < %v)", ff.Property, ff.Value))
		case "lte", "-lte", "<=":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s <= %v)", ff.Property, ff.Value))
		case "startsWith", "-startsWith":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | startswith(\"%v\"))", ff.Property, ff.Value))
		case "endsWith", "-endsWith":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | endswith(\"%v\"))", ff.Property, ff.Value))
		case "datelt", "-datelt":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | datelt(\"%v\"))", ff.Property, ff.Value))
		case "datelte", "-datelte", "olderthan", "-olderthan":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | datelte(\"%v\"))", ff.Property, ff.Value))
		case "dategt", "-dategt":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | dategt(\"%v\"))", ff.Property, ff.Value))
		case "dategte", "-dategte", "newerthan", "-newerthan":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | dategte(\"%v\"))", ff.Property, ff.Value))
		case "version", "-version":
			queryParts = append(queryParts, fmt.Sprintf("select(.%s | version(\"%v\"))", ff.Property, ff.Value))
		}
	}

	// Select
	if !f.HasPostJQFilter() {
		selectStatement := []string{}
		deleteStatements := []string{}

		properties := []string{}
		for _, pluck := range f.Pluck {
			properties = append(properties, strings.Split(pluck, ",")...)
		}

		for _, pluck := range properties {
			// Support property renaming
			alias, property, found := strings.Cut(pluck, ":")
			if !found {
				property = alias
			}
			switch {
			case property == "**":
				// Do nothing (select everything)
			case strings.HasPrefix(property, "!"):
				// Remove specific properties
				deleteStatements = append(deleteStatements, fmt.Sprintf("del(.%s)", property[1:]))
			case property != "":
				if f.AsCSV || f.AsTSV {
					selectStatement = append(selectStatement, fmt.Sprintf(".%s", property))
				} else {
					if strings.Contains(property, ".") && !f.Flatten {
						parts := strings.Split(property, ".")

						// FIXME: this only works for max of two nested items
						// {foo:{bar: .foo.bar}}
						selectStatement = append(selectStatement, fmt.Sprintf("%s:{%s:.%s}", parts[0], parts[1], strings.Join(parts[0:2], ".")))
					} else {
						// selectStatement = append(selectStatement, pluck)
						selectStatement = append(selectStatement, fmt.Sprintf(`"%s":.%s`, alias, property))
					}
				}
			}
		}

		// Apply delete statements before queries
		queryParts = append(queryParts, deleteStatements...)

		if len(selectStatement) > 0 {
			if f.AsCSV || f.AsTSV {
				queryParts = append(queryParts, fmt.Sprintf("[%s]", strings.Join(selectStatement, ",")))
			} else {
				queryParts = append(queryParts, fmt.Sprintf("{%s}", strings.Join(selectStatement, ",")))
			}
		}
	}

	// Post processing
	if f.PostJQ != "" {
		queryParts = append(queryParts, f.PostJQ)
	}

	if f.AsCSV {
		// FIXME: Use quotes on csv output, e.g. ignore the sub command
		queryParts = append(queryParts, `@csv|sub("\"";"";"g")`)
	}

	if f.AsTSV {
		queryParts = append(queryParts, "@tsv")
	}

	return strings.Join(queryParts, "|")
}

func GetKeys(v any) ([]string, error) {
	q, err := gojq.Parse(".|paths|join(\".\")")
	if err != nil {
		return nil, err
	}
	iter := q.RunWithContext(context.Background(), v)

	keys := []string{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		switch val := v.(type) {
		case string:
			keys = append(keys, val)
		}
	}
	return keys, nil
}

func (f JSONFilters) ApplyQuery(data any, query string) ([]interface{}, error) {
	if !strings.HasPrefix(query, ".") {
		query = "." + query + "[]"
	}
	q, err := gojq.Parse(query)

	if err != nil {
		return nil, err
	}

	iter := q.RunWithContext(context.Background(), data)
	output := make([]interface{}, 0)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		output = append(output, v)
	}

	return output, nil
}

func (f JSONFilters) filterJSONUsingJQ(data any, property string, showHeaders bool, setHeaderFunc func(string)) ([]byte, error) {

	rawJQQuery := f.GetJQQuery(property)
	f.Logger.Infof("jq query: %s", rawJQQuery)

	query, err := gojq.Parse(rawJQQuery)
	if err != nil {
		return nil, err
	}

	code, err := gojq.Compile(
		query,
		gojq.WithFunction("version", 1, 1, extStringFunc(matchVersionConstraint)),
		gojq.WithFunction("datelt", 1, 1, extStringFunc(dateOlderThan)),
		gojq.WithFunction("datelte", 1, 1, extStringFunc(dateOlderThanEqual)),
		gojq.WithFunction("dategt", 1, 1, extStringFunc(dateNewerThan)),
		gojq.WithFunction("dategte", 1, 1, extStringFunc(dateNewerThanEqual)),
	)
	if err != nil {
		return nil, err
	}

	iter := code.RunWithContext(context.Background(), data)
	var output bytes.Buffer
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		if vstr, ok := v.(string); ok {
			output.Write([]byte(vstr))
			output.WriteByte('\n')
		} else {
			if out, err := json.Marshal(v); err == nil {
				output.Write(out)
				output.WriteByte('\n')
			}
		}
	}

	setHeaderFunc(strings.Join(f.Pluck, ","))
	return output.Bytes(), nil
}

func (f JSONFilters) filterJSON(jsonValue string, property string, showHeaders bool, setHeaderFunc func(string)) ([]byte, error) {
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
	} else if v.IsArray() {
		jq = gojsonq.New().FromString(v.String())
	} else {
		return []byte(v.Str), nil
	}

	// Add custom filters
	jq.Macro("like", matchWithWildcards)
	jq.Macro("-like", matchWithWildcards)
	jq.Macro("-notlike", matchWithWildcardsNegated)
	jq.Macro("notlike", matchWithWildcardsNegated)

	jq.Macro("match", matchWithRegex)
	jq.Macro("-match", matchWithRegex)
	jq.Macro("-notmatch", matchWithRegexNegated)
	jq.Macro("notmatch", matchWithRegexNegated)

	// date filters
	jq.Macro("datelt", dateOlderThan)
	jq.Macro("datelte", dateOlderThanEqual)
	jq.Macro("olderthan", dateOlderThanEqual)

	jq.Macro("dategt", dateNewerThan)
	jq.Macro("dategte", dateNewerThanEqual)
	jq.Macro("newerthan", dateNewerThanEqual)

	// Version filters
	jq.Macro("version", matchVersionConstraint)

	for _, query := range f.Filters {
		Logger.Debugf("filtering data: %s %s %s", query.Property, query.Operation, query.Value)
		jq.Where(query.Property, query.Operation, query.Value)
	}

	if errs := jq.Errors(); len(errs) > 0 {
		Logger.Warnf("filter errors. %v", errs)
	}

	Logger.Debugf("Pluck values: %v", f.Pluck)
	// format values (using gjson)
	// skip flatten and select if a only a globstar is provided
	// selectAllProperties := len(f.Pluck) == 1 && f.Pluck[0] == "**"
	// && !selectAllProperties
	if (len(f.Pluck) > 0) || f.Flatten {
		var bsub bytes.Buffer
		jq.Writer(&bsub)
		formattedJSON := gjson.ParseBytes(bsub.Bytes())

		if formattedJSON.IsArray() {
			outputValues := make([]string, 0)

			if showHeaders {
				outputValues = append(outputValues, expandHeaderProperties(&formattedJSON, f.Pluck))
			}

			for _, myval := range formattedJSON.Array() {

				if myval.IsObject() {
					if line, keys := f.pluckJsonValues(&myval, f.Pluck); line != "" {
						outputValues = append(outputValues, line)
						setHeaderFunc(strings.Join(keys, ","))
					}
				} else {
					outputValues = append(outputValues, myval.Raw)
				}
			}
			return []byte(strings.Join(outputValues, "\n")), formatErrors(jq.Errors())
		}

		if line, keys := f.pluckJsonValues(&formattedJSON, f.Pluck); line != "" {
			setHeaderFunc(strings.Join(keys, ","))
			return []byte(line), formatErrors(jq.Errors())
		}

		Logger.Debugf("ERROR: gjson path does not exist. %v", f.Pluck)
		return []byte(""), formatErrors(jq.Errors())
	}

	jq.Writer(&b)

	// Convert back to an object if it
	if convertBackFromArray {
		return removeJSONArrayValues(b.Bytes()), formatErrors(jq.Errors())
	}
	return b.Bytes(), formatErrors(jq.Errors())
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

func (f JSONFilters) pluckJsonValues(item *gjson.Result, properties []string) (string, []string) {
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
	if f.AsCSV {
		return convertToCSV(flatMap, flatKeys, ","), flatKeys
	} else if f.AsTSV {
		return convertToCSV(flatMap, flatKeys, "\t"), flatKeys
	} else if f.AsCompletionFormat {
		return convertToLine(flatMap, flatKeys), flatKeys
	} else if f.Flatten {
		v, err = json.Marshal(flatMap)
	} else {
		// unflatten
		Logger.Debugf("running unflatten. %v", pathPatterns)
		if len(pathPatterns) == 1 && pathPatterns[0] == "**" {
			Logger.Debugf("Returning all keys because globstar is being used")
			return item.Raw, flatKeys
		}

		// Protect against large amount of keys adn
		maxKeyCount := int64(10000)
		keyCount := int64(len(flatMap))
		if keyCount > maxKeyCount {
			if f.Logger != nil {
				itemID := ""
				if v := item.Get("id"); v.Exists() {
					itemID = v.Str
				}
				f.Logger.Warnf("Detected json with a large number of keys, returning all data by default. Use jq for further filtering. total_keys=%d, id=%s", keyCount, itemID)
			}

			return item.Raw, flatKeys
		}

		v, err = flatten.UnflattenOrdered(flatMap, flatKeys)
		Logger.Debugf("Finished unflatten")
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

func convertToCSV(flatMap map[string]interface{}, keys []string, separator string) string {
	buf := bytes.Buffer{}
	if separator == "" {
		separator = ","
	}
	for i, key := range keys {
		if i != 0 {
			// handle for empty non-existent values by leaving it blank
			buf.WriteString(separator)
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

func convertToLine(flatMap map[string]interface{}, keys []string) string {
	buf := bytes.Buffer{}
	for i, key := range keys {
		if i != 0 {
			// handle for empty non-existent values by leaving it blank
			if i == 1 {
				buf.WriteString("\t")
			} else {
				buf.WriteString(" | ")
			}
		}
		if value, ok := flatMap[key]; ok {
			if marshalledValue, err := json.Marshal(value); err != nil {
				Logger.Warningf("failed to marshal value. value=%v, err=%s", value, err)
			} else {
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
	}
	return buf.String()
}

func dateNewerThan(x, y interface{}) (bool, error) {
	rawDateA, okx := x.(string)
	rawDateB, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	var err error
	dateA, err := timestamp.ParseTimestamp(rawDateA)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	dateB, err := timestamp.ParseTimestamp(rawDateB)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	return dateB.UnixNano() < dateA.UnixNano(), nil
}

func dateNewerThanEqual(x, y interface{}) (bool, error) {
	rawDateA, okx := x.(string)
	rawDateB, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	var err error
	dateA, err := timestamp.ParseTimestamp(rawDateA)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	dateB, err := timestamp.ParseTimestamp(rawDateB)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	return dateB.UnixNano() <= dateA.UnixNano(), nil
}

func dateOlderThan(x, y interface{}) (bool, error) {
	rawDateA, okx := x.(string)
	rawDateB, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	var err error
	dateA, err := timestamp.ParseTimestamp(rawDateA)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	dateB, err := timestamp.ParseTimestamp(rawDateB)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	return dateB.UnixNano() > dateA.UnixNano(), nil
}

func dateOlderThanEqual(x, y interface{}) (bool, error) {
	rawDateA, okx := x.(string)
	rawDateB, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	var err error
	dateA, err := timestamp.ParseTimestamp(rawDateA)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	dateB, err := timestamp.ParseTimestamp(rawDateB)
	if err != nil {
		return false, fmt.Errorf("only date strings are supported. %w", err)
	}

	return dateB.UnixNano() >= dateA.UnixNano(), nil
}

func matchWithWildcards(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	return matcher.MatchWithWildcards(xs, pattern)
}

func matchWithWildcardsNegated(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	match, err := matcher.MatchWithWildcards(xs, pattern)
	return !match, err
}

func matchWithRegex(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	return matcher.MatchWithRegex(xs, pattern)
}

func matchWithRegexNegated(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	pattern, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("wildcard matching only supports strings")
	}

	match, err := matcher.MatchWithRegex(xs, pattern)
	return !match, err
}

func toString(v any) string {
	switch v := v.(type) {
	case string:
		return v
	}
	return ""
}

func extStringFunc(f func(x, y interface{}) (bool, error)) func(x any, xs []any) any {
	return func(x any, xs []any) any {
		match, err := f(toString(x), toString(xs[0]))
		if err != nil {
			return err
		}
		if match {
			return x
		}
		return nil
	}
}

func matchVersionConstraint(x, y interface{}) (bool, error) {
	xs, okx := x.(string)
	ys, oky := y.(string)
	if !okx || !oky {
		return false, fmt.Errorf("version matching only supports strings")
	}

	currentVersion, err := version.NewVersion(xs)
	if err != nil {
		// Treat invalid versions as 0.0.0
		currentVersion = version.Must(version.NewVersion("0.0.0"))
	}

	constraint, err := version.NewConstraint(strings.ReplaceAll(ys, ",", ", "))
	if err != nil {
		return false, fmt.Errorf("invalid version constraint. %w", err)
	}

	return constraint.Check(currentVersion), nil
}
