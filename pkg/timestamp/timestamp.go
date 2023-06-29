package timestamp

import (
	"fmt"
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

func TryGetTimestamp(value string, encode bool, utc bool) (string, error) {
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