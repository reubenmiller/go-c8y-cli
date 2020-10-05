package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/reubenmiller/go-c8y/pkg/c8y"
)

// MustParseJSON parses a string and returns the map structure
func MustParseJSON(value string) map[string]interface{} {
	data := make(map[string]interface{})

	if isJSONString(value) {
		if err := parseJSONStructure(value, data); err != nil {
			panic(errors.Wrap(err, "Invalid JSON"))
		}
		return data
	}

	if err := parseShorthandJSONStructure(value, data); err != nil {
		panic(errors.Wrap(err, "Invalid shorthand JSON"))
	}
	return data
}

func isJSONString(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

func hasQuotes(value string) bool {
	return (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'"))
}

func isNumber(value string) (float64, bool) {
	if value == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(value, 64)
	return f, err == nil
}

func isArray(value string) ([]string, bool) {
	if strings.HasPrefix(value, "[") && strings.HasPrefix(value, "[") {
		return strings.Split(value[1:len(value)-1], ","), true
	}
	return []string{}, false
}

// parseJSON converts either a
func parseJSONStructure(value string, data map[string]interface{}) error {
	rdr := strings.NewReader(value)
	if err := c8y.DecodeJSONReader(rdr, &data); err != nil {
		return errors.Wrap(err, "invalid json")
	}
	return nil
}

//
// Examples:
// "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={'id': '12345'}"
// "text=one,severity=MAJOR,type=test_Type,time=2019-01-01,source={id: '12345'}"
//	->	{"severity":"MAJOR","source":{"id":"12345"},"text":"one","time":"2019-01-01","type":"test_Type"}
func parseValue(value string) interface{} {
	propValue := strings.TrimSpace(value)

	if isJSONString(propValue) {
		// Add quotes to keys
		re := regexp.MustCompile(`["']?(\w+)["']?\s*:`)
		propValue = re.ReplaceAllString(propValue, "\"${1}\":")

		// TODO: allow use of single quotes as double quotes
		propValue = strings.ReplaceAll(propValue, "'", "\"")

		jsonMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(propValue), &jsonMap); err != nil {
			Logger.Warningf("Invalid json. %s", err)

			// Try parsing
			return parseValue(propValue[1 : len(propValue)-1])
		}
		return jsonMap
	} else if values, valid := isArray(propValue); valid {
		// parse array values
		valueArray := []interface{}{}
		for _, item := range values {
			Logger.Debugf("item: %s", item)
			valueArray = append(valueArray, parseValue(item))
		}
		return valueArray
	} else if f, valid := isNumber(propValue); valid && !hasQuotes(propValue) {
		// value is a number (int or float)
		return f
	} else if propValue == "true" {
		return true
	} else if propValue == "false" {
		return false
	} else {
		if hasQuotes(propValue) {
			// remove quotes
			propValue = propValue[1 : len(propValue)-1]
		}
		return propValue
	}
}

// parseStructure splits a flat comma separated list to a json structure
// values := "key1=value1,key2=value2,key3=value3"
// https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-shorthand.html
func parseShorthandJSONStructure(value string, data map[string]interface{}) error {
	validItems := 0

	valuePairs := strings.Split(value, "=")

	if len(value) > 0 {
		Logger.Debugf("Input: %s", value)
	}

	outputValues := []string{}
	for _, item := range valuePairs {
		if strings.ContainsAny(item, "]}") {
			if strings.HasSuffix(item, "]") || strings.HasSuffix(item, "}") {
				// Last value
				outputValues = append(outputValues, item)
			} else {
				if pos := strings.LastIndex(item, ","); pos > -1 {
					outputValues = append(outputValues, item[0:pos], item[pos+1:])
				}
			}

		} else if strings.Contains(item, ",") {
			outputValues = append(outputValues, strings.Split(item, ",")...)
		} else {
			outputValues = append(outputValues, item)
		}
	}

	if value == "" {
		return nil
	}

	if len(outputValues)%2 != 0 {
		panic("Uneven number of key value pairs")
	}

	for i := 0; i < len(outputValues); i += 2 {
		key := strings.Trim(outputValues[i], " ")
		data[key] = parseValue(outputValues[i+1])
		validItems++
	}

	Logger.Debugf("Output: %v", outputValues)

	if validItems == 0 {
		return fmt.Errorf("Input contains no valid shorthand data")
	}

	return nil
}

// GetFileContentType TODO: Fix mime detection because it currently returns only application/octet-stream
func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

// getTempFilepath returns a temp file path. If outputDir is empty, then a temp folder will be created
func getTempFilepath(name string, outputDir string) (string, error) {
	directory := "./"

	if outputDir == "" {
		tempDir, err := ioutil.TempDir("", "go-c8y_")

		if err != nil {
			return "", fmt.Errorf("Could not create temp folder. %s", err)
		}
		directory = tempDir
	}

	return path.Join(directory, name), nil
}

// saveResponseToFile saves a response to file
// @filename	filename
// @directory	output directory. If empty, then a temp directory will be used
// if filename
func saveResponseToFile(resp *c8y.Response, filename string, append bool) (string, error) {

	var out *os.File
	var err error
	if append {
		out, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		out, err = os.Create(filename)
	}

	if err != nil {
		return "", fmt.Errorf("Could not create file. %s", err)
	}
	defer out.Close()

	// Writer the body to file
	Logger.Printf("header: %v", resp.Header)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy file contents to file. %s", err)
	}

	if fullpath, err := filepath.Abs(filename); err == nil {
		return fullpath, nil
	}
	return filename, nil
}

func getSessionHomeDir() string {
	var outputDir string

	if v := os.Getenv("C8Y_SESSION_HOME"); v != "" {
		outputDir = v
	} else {
		outputDir, _ = homedir.Dir()
		outputDir = filepath.Join(outputDir, ".cumulocity")
	}
	return outputDir
}
