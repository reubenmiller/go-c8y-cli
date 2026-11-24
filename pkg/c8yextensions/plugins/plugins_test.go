package plugins

import (
	"encoding/json"
	"testing"
)

func Test_hasFields(t *testing.T) {
	for body, want := range map[string]bool{
		``:              false,
		`{}`:            false,
		`null`:          false,
		`{"config":{}}`: true,
		`{"name":"x"}`:  true,
		`[1]`:           false,
	} {
		if got := hasFields(json.RawMessage(body)); got != want {
			t.Errorf("hasFields(%q) = %v, want %v", body, got, want)
		}
	}
}
