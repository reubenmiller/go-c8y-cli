package numbers

import "testing"

// TestRawNumberPreservesSource pins the "none" number format to the source
// representation: re-marshalling the parsed float would switch small/large
// magnitudes to scientific notation (0.000000000001 -> 1e-12), which the table
// output tests forbid.
func TestRawNumberPreservesSource(t *testing.T) {
	rn := &RawNumber{}
	cases := []struct {
		v    float64
		raw  string
		want string
	}{
		{0.000000000001, "0.000000000001", "0.000000000001"},
		{2038373478888889, "2038373478888889", "2038373478888889"},
		{0.000001, "0.000001", "0.000001"},
		{10, "10", "10"},
		{1234567.501202, "1234567.501202", "1234567.501202"},
	}
	for _, c := range cases {
		if got := rn.Display(c.v, c.raw, ""); got != c.want {
			t.Errorf("Display(%v, %q) = %q, want %q", c.v, c.raw, got, c.want)
		}
	}
}
