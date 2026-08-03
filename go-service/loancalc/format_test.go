package loancalc

import "testing"

func TestIndianGrouping(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{999, "999"},
		{1000, "1,000"},
		{99999, "99,999"},
		{100000, "1,00,000"},
		{644817, "6,44,817"},
		{5254344, "52,54,344"},
		{21042674, "2,10,42,674"},
		{30000000, "3,00,00,000"},
		{360181136, "36,01,81,136"},
		{-644817, "-6,44,817"},
	}
	for _, c := range cases {
		if got := commas(c.in); got != c.want {
			t.Errorf("commas(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
