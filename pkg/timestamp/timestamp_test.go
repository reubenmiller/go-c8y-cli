package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvalidDates(t *testing.T) {

	timestamp, err := ParseDurationRelativeToNow("2020010101")

	if err == nil {
		t.Errorf("Timestamp should throw an error. got %s, expected nil", err)
	}

	if timestamp != nil {
		t.Errorf("Timestamp should be nil. got=%v", timestamp)
	}
}

func TestMixingDatesWithRelativeTimeOffsets(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "subtract 1 hour with spaces around operator",
			input: "'2025-03-03T12:11:22.675221+01:00' - 1h",
			want:  "2025-03-03T11:11:22.675221+01:00",
		},
		{
			// date-only input has no timezone, so dateparse treats it as UTC midnight
			name:  "date-only string subtract 1 hour",
			input: "'2025-03-03' - 1h",
			want:  "2025-03-02T23:00:00Z",
		},
		{
			name:  "add 1 hour with spaces around operator",
			input: "'2025-03-03T12:11:22.675221+01:00' + 1h",
			want:  "2025-03-03T13:11:22.675221+01:00",
		},
		{
			name:  "subtract 1 hour without spaces",
			input: "'2025-03-03T12:11:22.675221+01:00' -1h",
			want:  "2025-03-03T11:11:22.675221+01:00",
		},
		{
			name:  "add 1 hour without spaces",
			input: "'2025-03-03T12:11:22.675221+01:00' +1h",
			want:  "2025-03-03T13:11:22.675221+01:00",
		},
		{
			name:  "add 1 day",
			input: "'2025-03-03T12:11:22.675221+01:00' + 1d",
			want:  "2025-03-04T12:11:22.675221+01:00",
		},
		{
			name:  "subtract combined duration hours and minutes",
			input: "'2025-03-03T12:11:22.675221+01:00' - 1h30m",
			want:  "2025-03-03T10:41:22.675221+01:00",
		},
		{
			name:  "quoted date with no offset",
			input: "'2025-03-03T12:11:22.675221+01:00'",
			want:  "2025-03-03T12:11:22.675221+01:00",
		},
		{
			name:  "quoted date with UTC timezone",
			input: "'2025-06-15T00:00:00Z' + 2h",
			want:  "2025-06-15T02:00:00Z",
		},
		{
			name:  "quoted date subtract crossing midnight",
			input: "'2025-03-03T01:00:00+01:00' - 2h",
			want:  "2025-03-02T23:00:00+01:00",
		},
		{
			name:    "invalid date inside quotes",
			input:   "'not-a-date' + 1h",
			wantErr: true,
		},
		{
			name:    "invalid offset after valid date",
			input:   "'2025-03-03T12:11:22.675221+01:00' * 1h",
			wantErr: true,
		},
		{
			// Unix timestamp (seconds since epoch) + offset; base is always UTC
			name:  "unix timestamp add 1 hour",
			input: "1773403680 + 1h",
			want:  time.Unix(1773403680, 0).UTC().Add(time.Hour).Format(time.RFC3339Nano),
		},
		{
			name:  "unix timestamp subtract 30 minutes",
			input: "1773403680 - 30m",
			want:  time.Unix(1773403680, 0).UTC().Add(-30 * time.Minute).Format(time.RFC3339Nano),
		},
		{
			name:  "unix timestamp add 1 hour no spaces",
			input: "1773403680+1h",
			want:  time.Unix(1773403680, 0).UTC().Add(time.Hour).Format(time.RFC3339Nano),
		},
		{
			// "+ -1d" is mathematically equivalent to "- 1d"
			name:  "unix timestamp plus negative duration",
			input: "1773403680 + -1d",
			want:  time.Unix(1773403680, 0).UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano),
		},
		{
			// "- -1d" is mathematically equivalent to "+ 1d"
			name:  "unix timestamp minus negative duration",
			input: "1773403680 - -1d",
			want:  time.Unix(1773403680, 0).UTC().Add(24 * time.Hour).Format(time.RFC3339Nano),
		},
		{
			name:  "quoted date plus negative duration",
			input: "'2025-03-03T12:00:00Z' + -1h",
			want:  "2025-03-03T11:00:00Z",
		},
		{
			name:  "quoted date minus negative duration",
			input: "'2025-03-03T12:00:00Z' - -1h",
			want:  "2025-03-03T13:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := TryGetTimestamp(tt.input, false, false)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, out)
			}
		})
	}
}
