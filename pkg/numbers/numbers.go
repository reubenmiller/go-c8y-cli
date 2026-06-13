package numbers

import (
	"encoding/json"

	"github.com/dustin/go-humanize"
)

type NumberFormatter interface {
	Display(float64, string, string) string
}

type RawNumber struct{}

func (rawNum *RawNumber) Display(v float64, raw string, unit string) string {
	// Normalise the number the same way the v1 output path did (it decoded
	// and re-marshalled JSON before rendering), so e.g. 0.000000000001 is
	// shown as 1e-12 regardless of how it was written in the source document.
	// The raw text is preserved as a fallback for values json cannot encode.
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return raw
}

type NumberWithCommas struct{}

func (rawNum *NumberWithCommas) Display(v float64, raw string, unit string) string {
	return humanize.Commaf(v)
}

type NumberViewOptions struct {
	Precision        int
	ActivateRangeMin float64
	ActivateRangeMax float64
}

func NewNumberViewOptions(precision int, rangeMin float64, rangeMax float64) *NumberViewOptions {
	return &NumberViewOptions{
		Precision:        precision,
		ActivateRangeMin: rangeMin,
		ActivateRangeMax: rangeMax,
	}
}

func (num *NumberViewOptions) GetPrecision() int {
	if num.Precision < 0 {
		return 2
	}
	return num.Precision
}

func (num *NumberViewOptions) Display(v float64, raw string, unit string) string {
	if num.IsVerySmallOrVeryLarge(v) {
		return humanize.SIWithDigits(v, num.Precision, unit)
	}
	return raw
}

func (num *NumberViewOptions) IsVerySmallOrVeryLarge(v float64) bool {
	return v <= num.ActivateRangeMin || v >= num.ActivateRangeMax
}
