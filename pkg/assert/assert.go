package assert

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func EqualMarshalJSON(t *testing.T, got interface{}, want string) {
	out, err := json.Marshal(got)
	OK(t, err)

	if !bytes.EqualFold(out, []byte(want)) {
		t.Errorf(`Marshal. wanted=%s, got=%s`, want, got)
	}
}

func EqualJSON(t *testing.T, got []byte, want string) {
	if !bytes.EqualFold(got, []byte(want)) {
		t.Errorf(`Marshal. wanted=%s, got=%s`, want, got)
	}
}

func OK(t *testing.T, err error) {
	if err != nil {
		t.Errorf(`Error. wanted=nil, got=%s`, err)
	}
}

func True(t *testing.T, value bool) {
	if !value {
		t.Errorf(`Error. wanted=%v, got=%v`, true, value)
	}
}

func False(t *testing.T, value bool) {
	if value {
		t.Errorf(`Error. wanted=%v, got=%v`, false, value)
	}
}

func ErrorType(t *testing.T, got error, want error) {
	if !errors.Is(got, want) {
		t.Errorf(`Error. wanted=%s, got=%s`, want, got)
	}

}

func ErrorEqual(t *testing.T, got error, err error) {
	if !errors.Is(got, err) {
		t.Errorf(`Error. wanted=nil, got=%s`, err)
	}
}

func ErrorNil(t *testing.T, got error) {
	if got != nil {
		t.Errorf(`Error. wanted=nil, got=%s`, got)
	}
}
