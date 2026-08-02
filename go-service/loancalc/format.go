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
	var out string
	rem := len(s) % 3
	if rem > 0 {
		out = s[:rem]
	}
	for i := rem; i < len(s); i += 3 {
		if out != "" {
			out += ","
		}
		out += s[i : i+3]
	}
	return neg + out
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
