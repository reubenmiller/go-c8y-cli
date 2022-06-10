package flags

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/reubenmiller/go-c8y-cli/v2/pkg/timestamp"
	"github.com/spf13/cobra"
)

// ErrInvalidDuration invalid duration
var ErrInvalidDuration = errors.New("invalid duration")

// GetDuration get duration with an option of assuming the string is referring to seconds
func GetDuration(v string, inferUnit bool, unit time.Duration) (time.Duration, error) {

	// Accept more duration formations like "30d" which is not natively supported by go
	if d, err := timestamp.ParseDuration(v); err == nil {
		return d, nil
	}

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

type DurationGenerator func(time.Duration) time.Duration

// GetDurationGenerator returns a random duration generator func. The generator will return a random duration between the given min or max
func GetDurationGenerator(cmd *cobra.Command, minFlag, maxFlag string, inferUnit bool, unit time.Duration) (DurationGenerator, error) {
	minDuration, err := GetDurationFlag(cmd, minFlag, inferUnit, unit)
	if err != nil {
		return nil, err
	}
	maxDuration, err := GetDurationFlag(cmd, maxFlag, inferUnit, unit)
	if err != nil {
		return nil, err
	}
	min := int64(minDuration)
	max := int64(maxDuration)

	generator := func(fixed time.Duration) (delay time.Duration) {
		if max > 0 && max > min {
			// Note: max must be > min otherwise rand.Int63n throws an error!
			delay = time.Duration(rand.Int63n(max-min) + min)
		} else if min > 0 {
			delay = time.Duration(min)
		}
		if delay <= 0 {
			delay = fixed
		}
		return
	}
	return generator, nil
}
