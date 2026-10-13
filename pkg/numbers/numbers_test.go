package numbers

import "testing"

func TestNumberViewOptionsDisplay(t *testing.T) {
	opts := NewNumberViewOptions(2, 0.001, 1e9)
	cases := []struct {
		value float64
		raw   string
		unit  string
		want  string
	}{
		{0, "0", "", "0"},
		{0, "0", "V", "0 V"},
		{0.00001, "0.00001", "", "10 µ"},
		{1e12, "1000000000000", "", "1 T"},
		{10, "10", "", "10"},
	}
	for _, c := range cases {
		if got := opts.Display(c.value, c.raw, c.unit); got != c.want {
			t.Errorf("Display(%v, %q, %q) = %q, want %q", c.value, c.raw, c.unit, got, c.want)
		}
	}
}
