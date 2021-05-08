package flags

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// ErrInvalidDuration invalid duration
var ErrInvalidDuration = errors.New("invalid duration")

// GetDuration get duration with an option of assuming the string is referring to seconds
func GetDuration(v string, inferUnit bool, unit time.Duration) (time.Duration, error) {
	d, err := time.ParseDuration(v)
	if err == nil {
		return d, nil
	}

	if !inferUnit {
		return 0, fmt.Errorf("%w: %s", ErrInvalidDuration, err)
	}

	// Infer type (i.e. when no unit is specified)
	rawValue, err := strconv.ParseFloat(v, 64)
	if err == nil {
		if unit == 0 {
			return 0, nil
		}
		// s := rawValue / float64(unit)
		return time.Duration(rawValue) * unit, nil
	}
	return 0, ErrInvalidDuration
}

func GetDurationFlag(cmd *cobra.Command, name string, inferUnit bool, unit time.Duration) (time.Duration, error) {
	rawValue, err := cmd.Flags().GetString(name)
	if err != nil {
		return time.Duration(0), err
	}
	return GetDuration(rawValue, inferUnit, unit)
}
