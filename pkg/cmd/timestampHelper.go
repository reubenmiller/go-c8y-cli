package cmd

import (
	"time"

	"github.com/karrick/tparse/v2"
)

// getTimestampUsingOffset returns a timestamp relative to now
// Examples
// getTimeRelativeToNow("-1m")
func parseDurationRelativeToNow(offsetDuration string) (*time.Time, error) {
	now := time.Now()
	return getTimestampUsingOffset(now, offsetDuration)
}

// getTimestampUsingOffset returns a timestamp relative to a base timestamp
// example: +1d3w4mo-7y6h4m
// TODO: an offsetDuration of "30" throws a panic!
func getTimestampUsingOffset(now time.Time, offsetDuration string) (*time.Time, error) {
	another, err := tparse.AddDuration(now, offsetDuration)
	if err != nil {
		return nil, err
	}
	return &another, nil
}
