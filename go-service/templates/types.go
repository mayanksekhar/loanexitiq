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
	Outstanding  float64
	SelectedIdx  int
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

func f2(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

func formatINRf(n float64) string {
	return formatINR(n)
}

func calloutLead(month int, paid, cleared float64) string {
	return "By month " + strconv.Itoa(month) + " you have paid " + formatINR(paid) +
		". Only " + formatINR(cleared) + " of that reduced what you owe."
}

func calloutSub(pct int, outstanding, loan float64) string {
	return strconv.Itoa(pct) + "% went to interest, and you still owe " +
		formatINR(outstanding) + " of the original " + formatINR(loan) + "."
}

func lumpHeadline(adv float64) string {
	if adv <= 0 {
		return "Both routes cost about the same here."
	}
	return "Keeping your EMI the same and finishing sooner saves " + formatINR(adv) + " more than lowering your EMI, for the exact same part-payment."
}

func lumpFreeNote(lender string, pct, allowance, lump float64, isFree bool) string {
	if pct == 0 {
		return lender + " does not allow free part-payment at this point, so any lumpsum here may attract a charge. Check your sanction letter."
	}
	base := lender + " allows up to " + formatINR(allowance) + " of part-payment with no charge right now."
	if isFree {
		return base + " Your " + formatINR(lump) + " is within that, so it costs nothing to pay."
	}
	return base + " Your " + formatINR(lump) + " goes past that, so the excess may attract a fee."
}

func lumpMonthsLine(months int) string {
	return strconv.Itoa(months) + " months left"
}

func lumpSavedLine(saved float64) string {
	return "Interest saved: " + formatINR(saved)
}

func lumpFinishLine(months, saved int) string {
	return strconv.Itoa(months) + " months left, " + strconv.Itoa(saved) + " months sooner"
}
