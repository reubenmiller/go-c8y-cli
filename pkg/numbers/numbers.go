package numbers

import "github.com/dustin/go-humanize"

type NumberFormatter interface {
	Display(float64, string, string) string
}

type RawNumber struct{}

func (rawNum *RawNumber) Display(v float64, raw string, unit string) string {
	// "none" formatting shows the number exactly as it appeared in the source
	// document: 0.000000000001 stays 0.000000000001, not 1e-12. raw is the
	// original JSON token (node.Raw), so re-marshalling the parsed float — which
	// would switch large/small magnitudes to scientific notation — must be
	// avoided.
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
