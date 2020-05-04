package cmd

import (
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func formatC8yTimestamp(timestamp time.Time) string {
	return encodeC8yTimestamp(timestamp.Format(time.RFC3339Nano))
}

func encodeC8yTimestamp(value string) string {
	return strings.ReplaceAll(value, "+", "%2B")
}

func decodeC8yTimestamp(value string) string {
	return strings.ReplaceAll(value, "%2B", "+")
}

// tryGetTimestampFlag try to return the date time as either a relative time duration to now
//
// 1. Try parsing a relative time string
// 2. Try parsing the string as a date
// 3. Return the date as is (except by replacing any "+" with "%2B")
func tryGetTimestampFlag(cmd *cobra.Command, name string) (string, error) {
	val, err := cmd.Flags().GetString(name)

	if err != nil {
		return "", errors.Wrap(err, "could not read flag")
	}

	// Try parsing relative timestamp
	if ts, err := parseDurationRelativeToNow(val); err == nil {
		return formatC8yTimestamp(*ts), nil
	}

	// Try parsing timestamp (if valid)
	if timestamp, err := dateparse.ParseAny(val); err == nil {
		return formatC8yTimestamp(timestamp), nil
	}

	// Return the date without parsing it, just encode it. If error then cumulocity will return an error
	return encodeC8yTimestamp(val), nil
}
