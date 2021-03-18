package c8yfetcher

import (
	"testing"
)

func Test_parsingIDs(t *testing.T) {

	ids, names := parseAndSanitizeIDs([]string{"12345, 1231123"})

	if len(ids) != 2 {
		t.Errorf("Failed to pass ids. wanted=2, got=%d", len(ids))
	}

	if len(names) != 0 {
		t.Errorf("Failed to pass ids. wanted=0, got=%d", len(names))
	}
}

func Test_parsingIDsAndNames(t *testing.T) {
	ids, names := parseAndSanitizeIDs([]string{"12345, 1231123", "0124,my values"})

	if len(ids) != 3 {
		t.Errorf("Failed to pass ids. wanted=2, got=%d", len(ids))
	}

	if len(names) != 1 {
		t.Errorf("Failed to pass names. wanted=1, got=%d", len(names))
	}
	if names[0] != "my values" {
		t.Errorf("Invalid name parsing. wanted=my values, got=%s", names[0])
	}
}
