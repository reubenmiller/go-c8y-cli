package timestamp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/karrick/tparse/v2"
)

// ParseDurationRelativeToNow returns a timestamp relative to now
// Examples
// ParseDurationRelativeToNow("-1m")
func ParseDurationRelativeToNow(offsetDuration string) (*time.Time, error) {
	now := time.Now()
	if offsetDuration == "now" {
		return GetTimestampUsingOffset(now, "")
	}
	return GetTimestampUsingOffset(now, offsetDuration)
}

// GetTimestampUsingOffset returns a timestamp relative to a base timestamp
// example: +1d3w4mo-7y6h4m
func GetTimestampUsingOffset(now time.Time, offsetDuration string) (*time.Time, error) {
	another, err := tparse.AddDuration(now, offsetDuration)
	if err != nil {
		return nil, err
	}
	return &another, nil
}

// ParseDuration converts a duration string representation to a time.Duration
// The duration is in reference to now.
func ParseDuration(duration string) (time.Duration, error) {
	return tparse.AbsoluteDuration(time.Now(), duration)
}

func FormatC8yTimestamp(timestamp time.Time, encode bool) string {
	if encode {
		return EncodeC8yTimestamp(timestamp.Format(time.RFC3339Nano))
	}
	return timestamp.Format(time.RFC3339Nano)
}

func FormatC8yDate(timestamp time.Time, encode bool, layout string) string {
	if layout == "" {
		layout = "2006-02-01"
	}
	if encode {
		return EncodeC8yTimestamp(timestamp.Format(layout))
	}
	return timestamp.Format(layout)
}

func EncodeC8yTimestamp(value string) string {
	return strings.ReplaceAll(value, "+", "%2B")
}

func DecodeC8yTimestamp(value string) string {
	return strings.ReplaceAll(value, "%2B", "+")
}

// normalizeOffset combines an outer sign character ('+' or '-') with a raw
// duration string (which may itself start with '+' or '-') into a single
// tparse-compatible offset, applying standard sign arithmetic:
//
//	outer='+', inner='-' → '-'   outer='-', inner='-' → '+'
//
// Spaces within the duration part are also removed.
func normalizeOffset(outerSign string, durationPart string) string {
	durationPart = strings.ReplaceAll(strings.TrimSpace(durationPart), " ", "")
	if strings.HasPrefix(durationPart, "-") || strings.HasPrefix(durationPart, "+") {
		innerSign := string(durationPart[0])
		durationPart = durationPart[1:]
		if outerSign == innerSign {
			outerSign = "+"
		} else {
			outerSign = "-"
		}
	}
	return outerSign + durationPart
}

// parseQuotedDateWithOffset parses expressions of the form:
//
//	'<date>' [+|-] <duration>
//
// The date inside single quotes is parsed as an absolute timestamp and the
// optional trailing offset (e.g. "- 1h", "+2d") is applied with tparse.
// Returns an error if the value does not start with a single quote or if
// parsing fails.
func parseQuotedDateWithOffset(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "'") {
		return nil, fmt.Errorf("not a quoted date expression")
	}

	closeIdx := strings.Index(trimmed[1:], "'")
	if closeIdx == -1 {
		return nil, fmt.Errorf("unclosed single quote in date expression")
	}
	closeIdx++ // make relative to trimmed

	dateStr := trimmed[1:closeIdx]
	remainder := strings.TrimSpace(trimmed[closeIdx+1:])

	baseTime, err := dateparse.ParseAny(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", dateStr, err)
	}

	if remainder == "" {
		return &baseTime, nil
	}

	if !strings.HasPrefix(remainder, "+") && !strings.HasPrefix(remainder, "-") {
		return nil, fmt.Errorf("expected +/- offset after quoted date, got: %s", remainder)
	}

	// Normalise "- 1h 30m" / "+ -1d" → tparse-compatible offset
	offset := normalizeOffset(string(remainder[0]), remainder[1:])

	result, err := GetTimestampUsingOffset(baseTime, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid offset %q: %w", offset, err)
	}
	return result, nil
}

// parseUnixTimestampWithOffset parses expressions of the form:
//
//	<unix_seconds> +|- <duration>
//
// e.g. "1773403680 + 1h", "1773403680 -30m"
// A bare integer with no offset is intentionally not handled here so that
// values like YYYYMMDD fall through to dateparse unchanged.
func parseUnixTimestampWithOffset(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)

	// Scan leading digits
	end := 0
	for end < len(trimmed) && trimmed[end] >= '0' && trimmed[end] <= '9' {
		end++
	}
	if end == 0 {
		return nil, fmt.Errorf("not a unix timestamp expression")
	}

	remainder := strings.TrimSpace(trimmed[end:])

	// Require an explicit +/- offset — bare integers fall through to existing handling
	if remainder == "" || (!strings.HasPrefix(remainder, "+") && !strings.HasPrefix(remainder, "-")) {
		return nil, fmt.Errorf("not a unix timestamp with offset expression")
	}

	unixSec, err := strconv.ParseInt(trimmed[:end], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid unix timestamp %q: %w", trimmed[:end], err)
	}

	baseTime := time.Unix(unixSec, 0).UTC()

	// Normalise "- 1h 30m" / "+ -1d" → tparse-compatible offset
	offset := normalizeOffset(string(remainder[0]), remainder[1:])

	result, err := GetTimestampUsingOffset(baseTime, offset)
	if err != nil {
		return nil, fmt.Errorf("invalid offset %q: %w", offset, err)
	}
	return result, nil
}

func TryGetTimestamp(value string, encode bool, utc bool) (string, error) {
	// Try parsing quoted date with optional relative offset, e.g. '2025-03-03T12:00:00+01:00' - 1h
	if ts, err := parseQuotedDateWithOffset(value); err == nil {
		if utc {
			return FormatC8yTimestamp(ts.UTC(), encode), nil
		}
		return FormatC8yTimestamp(*ts, encode), nil
	} else if strings.HasPrefix(strings.TrimSpace(value), "'") {
		// Value looks like a quoted expression but failed to parse — surface the error
		return "", err
	}

	// Try parsing unix timestamp with offset, e.g. 1773403680 + 1h
	if ts, err := parseUnixTimestampWithOffset(value); err == nil {
		if utc {
			return FormatC8yTimestamp(ts.UTC(), encode), nil
		}
		return FormatC8yTimestamp(*ts, encode), nil
	}

	// Try parsing relative timestamp
	if ts, err := ParseDurationRelativeToNow(value); err == nil {
		if utc {
			return FormatC8yTimestamp(ts.UTC(), encode), nil
		}
		return FormatC8yTimestamp(*ts, encode), nil
	}

	// Try parsing timestamp (if valid)
	if timestamp, err := dateparse.ParseAny(value); err == nil {
		if utc {
			return FormatC8yTimestamp(timestamp.UTC(), encode), nil
		}
		return FormatC8yTimestamp(timestamp, encode), nil
	}

	if encode {
		// Return the date without parsing it, just encode it. If error then cumulocity will return an error
		return EncodeC8yTimestamp(value), nil
	}
	return value, nil
}

func TryGetDate(value string, encode bool, layout string) (string, error) {
	// Try parsing quoted date with optional relative offset
	if ts, err := parseQuotedDateWithOffset(value); err == nil {
		return FormatC8yDate(*ts, encode, layout), nil
	} else if strings.HasPrefix(strings.TrimSpace(value), "'") {
		return "", err
	}

	// Try parsing unix timestamp with offset, e.g. 1773403680 + 1h
	if ts, err := parseUnixTimestampWithOffset(value); err == nil {
		return FormatC8yDate(*ts, encode, layout), nil
	}

	// Try parsing relative date
	if ts, err := ParseDurationRelativeToNow(value); err == nil {
		return FormatC8yDate(*ts, encode, layout), nil
	}

	// Try parsing timestamp (if valid)
	if timestamp, err := dateparse.ParseAny(value); err == nil {
		return FormatC8yDate(timestamp, encode, layout), nil
	}

	if encode {
		// Return the date without parsing it, just encode it. If error then cumulocity will return an error
		return EncodeC8yTimestamp(value), nil
	}
	return value, nil
}

// ParseTimestamp parse a time stamp (accepts both relative and full timestamps)
func ParseTimestamp(value string) (ts time.Time, err error) {
	// Try parsing quoted date with optional relative offset
	if result, parseErr := parseQuotedDateWithOffset(value); parseErr == nil {
		ts = *result
		return
	} else if strings.HasPrefix(strings.TrimSpace(value), "'") {
		err = parseErr
		return
	}

	// Try parsing unix timestamp with offset, e.g. 1773403680 + 1h
	if result, parseErr := parseUnixTimestampWithOffset(value); parseErr == nil {
		ts = *result
		return
	}

	// Try parsing relative timestamp
	timestamp, err := ParseDurationRelativeToNow(value)

	if err == nil {
		ts = *timestamp
		return
	}

	// Try parsing timestamp (if valid)
	ts, err = dateparse.ParseAny(value)
	if err == nil {
		return
	}

	return
}

// AddDateTime adds an offset to a given timestamp (either as string or time.Time)
func AddDateTime(now any, offset string) (ts time.Time, err error) {
	// Try parsing relative timestamp

	var tsNow time.Time
	switch v := now.(type) {
	case string:
		tsNow, err = ParseTimestamp(v)
		if err != nil {
			return
		}
	case time.Time:
		tsNow = v
	case *time.Time:
		tsNow = *v
	default:
		err = fmt.Errorf("unsupported datetime type")
		return
	}

	ts1, offsetErr := GetTimestampUsingOffset(tsNow, offset)
	if offsetErr != nil {
		err = offsetErr
		return
	}
	ts = *ts1
	return
}
