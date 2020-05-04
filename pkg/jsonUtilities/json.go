package jsonUtilities

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var (
	objectPrefix = []byte("{")
	objectSuffix = []byte("}")

	arrayPrefix = []byte("[")
	arraySuffix = []byte("]")
)

// IsValidJSON returns true if the given byte array is a JSON array of JSON object
func IsValidJSON(v []byte) bool {
	val := bytes.TrimSpace(v)
	return json.Valid(val) && (IsJSONArray(val) || IsJSONObject(val))
}

// IsJSONArray returns true if the byte array represents a JSON array
func IsJSONArray(v []byte) bool {
	return bytes.HasPrefix(v, arrayPrefix) && bytes.HasSuffix(v, arraySuffix)
}

// IsJSONObject returns true if the byte array represents a JSON object
func IsJSONObject(v []byte) bool {
	return bytes.HasPrefix(v, objectPrefix) && bytes.HasSuffix(v, objectSuffix)
}

// UnescapeJSON replaces unicode escape characters with the actual character
func UnescapeJSON(v []byte) string {
	val, err := strconv.Unquote("\"" + strings.ReplaceAll(string(v), `"`, `\"`) + "\"")
	if err != nil {
		return string(v)
	}
	return val
}

type JsonDecodeError struct {
	Err error
}

var ErrOpenFile = errors.New("failed to open file")
var ErrJSONDecode = errors.New("failed to decode JSON")
var ErrReadFile = errors.New("failed to read file")

// DecodeJSONFile returns the contents of a json file as a map
func DecodeJSONFile(filename string) (map[string]interface{}, error) {
	var err error

	jsonFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, fmt.Errorf("%w", ErrOpenFile)
	}

	defer jsonFile.Close()

	contents := make(map[string]interface{})

	b, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrReadFile)
	}

	err = json.Unmarshal(b, &contents)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrJSONDecode)
	}
	return contents, nil
}
