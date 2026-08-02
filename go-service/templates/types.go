package templates

import (
	"fmt"
	"strconv"
)

type StatsData struct {
	EMI           float64
	InterestPaid  float64
	PrincipalPaid float64
	Outstanding   float64
	ExitMonth     int
	Tenure        int
}

func formatINR(n float64) string {
	return fmt.Sprintf("Rs %s", commas(int64(n)))
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

func ratePctStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

type LenderResult struct {
	Name         string
	Note         string
	RatePct      float64
	InterestPaid float64
	Fee          float64
	FeePct       float64
	Total        float64
	IsBest       bool
	IsWorst      bool
}

func donutSlice(part, total float64) string {
	if total <= 0 {
		return "0 251.33"
	}
	const circumference = 251.33 // 2 * pi * 40
	sliceLen := part / total * circumference
	if sliceLen < 0 {
		sliceLen = 0
	}
	remainder := circumference - sliceLen
	if remainder < 0 {
		remainder = 0
	}
	return strconv.FormatFloat(sliceLen, 'f', 2, 64) + " " + strconv.FormatFloat(remainder, 'f', 2, 64)
}

// note: LenderResult.FeePct added for single-bank view
