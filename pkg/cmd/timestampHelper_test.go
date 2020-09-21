package cmd

import (
	"testing"
)

func TestInvalidDates(t *testing.T) {

	timestamp, err := parseDurationRelativeToNow("2020010101")

	if err == nil {
		t.Errorf("Timestamp should throw an error. got %s, expected nil", err)
	}

	if timestamp != nil {
		t.Errorf("Timestamp should be nil. got=%v", timestamp)
	}
}
