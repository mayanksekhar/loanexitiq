package loancalc

import "strconv"

func formatINR(n float64) string {
	return "Rs " + commas(int64(n))
}

func commas(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := ""
	if n < 0 {
		neg = "-"
		s = s[1:]
	}
	if len(s) <= 3 {
		return neg + s
	}
	// Indian grouping: last three digits, then pairs.
	last3 := s[len(s)-3:]
	rest := s[:len(s)-3]
	out := ""
	for len(rest) > 2 {
		out = "," + rest[len(rest)-2:] + out
		rest = rest[:len(rest)-2]
	}
	if len(rest) > 0 {
		out = rest + out
	}
	return neg + out + "," + last3
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func ratePct(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64) + "%"
}
