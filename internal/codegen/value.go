// Package codegen generates the cobra CLI commands from the api/spec/json
// specifications. It is a port of the PowerShell scripts under
// scripts/build-cli and reproduces their output byte-for-byte.
package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Kind enumerates the JSON value kinds tracked by Value.
type Kind byte

// JSON value kinds
const (
	KindNull Kind = iota
	KindObject
	KindArray
	KindString
	KindNumber
	KindBool
)

// Value is an ordered, loosely typed JSON value. It mirrors how PowerShell's
// ConvertFrom-Json exposes the specification data: object keys keep their
// document order and member access is case-insensitive.
type Value struct {
	kind Kind
	keys []string
	obj  map[string]*Value
	arr  []*Value
	str  string
	num  json.Number
	b    bool
}

// ParseJSON decodes data into a Value preserving object key order.
func ParseJSON(data []byte) (*Value, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	v, err := parseValue(dec)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func parseValue(dec *json.Decoder) (*Value, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return parseToken(dec, tok)
}

func parseToken(dec *json.Decoder, tok json.Token) (*Value, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			v := &Value{kind: KindObject, obj: map[string]*Value{}}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyTok.(string)
				if !ok {
					return nil, fmt.Errorf("expected object key, got %v", keyTok)
				}
				child, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				if _, exists := v.obj[key]; !exists {
					v.keys = append(v.keys, key)
				}
				v.obj[key] = child
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return nil, err
			}
			return v, nil
		case '[':
			v := &Value{kind: KindArray}
			for dec.More() {
				child, err := parseValue(dec)
				if err != nil {
					return nil, err
				}
				v.arr = append(v.arr, child)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return nil, err
			}
			return v, nil
		}
		return nil, fmt.Errorf("unexpected delimiter %v", t)
	case string:
		return &Value{kind: KindString, str: t}, nil
	case json.Number:
		return &Value{kind: KindNumber, num: t}, nil
	case bool:
		return &Value{kind: KindBool, b: t}, nil
	case nil:
		return &Value{kind: KindNull}, nil
	}
	return nil, fmt.Errorf("unexpected token %v", tok)
}

// Get returns the named member of an object (case-insensitive, like
// PowerShell property access). Returns nil when missing or when v is not an
// object. Safe to call on a nil receiver.
func (v *Value) Get(key string) *Value {
	if v == nil || v.kind != KindObject {
		return nil
	}
	if m, ok := v.obj[key]; ok {
		return m
	}
	for k, m := range v.obj {
		if strings.EqualFold(k, key) {
			return m
		}
	}
	return nil
}

// Keys returns the object keys in document order.
func (v *Value) Keys() []string {
	if v == nil || v.kind != KindObject {
		return nil
	}
	return v.keys
}

// Items coerces the value to an array the way PowerShell's [array] cast does:
// nil stays empty, arrays return their elements and scalars become a
// single-element array.
func (v *Value) Items() []*Value {
	if v == nil || v.kind == KindNull {
		return nil
	}
	if v.kind == KindArray {
		return v.arr
	}
	return []*Value{v}
}

// IsNull reports whether the value is missing or JSON null.
func (v *Value) IsNull() bool {
	return v == nil || v.kind == KindNull
}

// Str renders the value the way PowerShell string interpolation does:
// null -> "", bool -> True/False, numbers use their literal form.
func (v *Value) Str() string {
	if v == nil {
		return ""
	}
	switch v.kind {
	case KindNull:
		return ""
	case KindString:
		return v.str
	case KindNumber:
		return v.num.String()
	case KindBool:
		if v.b {
			return "True"
		}
		return "False"
	case KindArray:
		parts := make([]string, 0, len(v.arr))
		for _, item := range v.arr {
			parts = append(parts, item.Str())
		}
		return strings.Join(parts, " ")
	}
	return ""
}

// Truthy reports the PowerShell boolean conversion of the value.
func (v *Value) Truthy() bool {
	if v == nil {
		return false
	}
	switch v.kind {
	case KindNull:
		return false
	case KindString:
		return v.str != ""
	case KindNumber:
		f, err := v.num.Float64()
		return err == nil && f != 0
	case KindBool:
		return v.b
	case KindArray:
		if len(v.arr) == 0 {
			return false
		}
		if len(v.arr) == 1 {
			return v.arr[0].Truthy()
		}
		return true
	case KindObject:
		return true
	}
	return false
}

// StrEquals reports a case-insensitive string comparison like PowerShell -eq.
func (v *Value) StrEquals(other string) bool {
	return strings.EqualFold(v.Str(), other)
}
