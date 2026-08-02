package loancalc

import "sort"

type FeeTier struct {
	UptoMonth int
	Pct       float64
}

func feeAtTiers(tiers []FeeTier, exitMonth int) float64 {
	for _, t := range tiers {
		if t.UptoMonth > 0 && exitMonth <= t.UptoMonth {
			return t.Pct
		}
	}
	for _, t := range tiers {
		if t.UptoMonth == 0 {
			return t.Pct
		}
	}
	if len(tiers) == 0 {
		return 0
	}
	return tiers[len(tiers)-1].Pct
}

type Lender struct {
	Name     string
	Note     string
	DRate    float64
	FeeTiers []FeeTier
}

func (l Lender) FeeAt(exitMonth int) float64 {
	return feeAtTiers(l.FeeTiers, exitMonth)
}

// Verified against published lender fee schedules, Aug 2026. See docs/SOURCES.md.
var Lenders = []Lender{
	{"ICICI LAP", "4% non-individual, part-pay free, 12mo clawback", 0,
		[]FeeTier{{0, 4}}},
	{"Kotak business", "4% fixed rate, unsecured business loan", 0,
		[]FeeTier{{0, 4}}},
	{"HDFC LAP", "up to 2.5% business use, 25%/yr part-pay free, MSME waiver", -0.1,
		[]FeeTier{{0, 2.5}}},
	{"SBI term loan", "3% only within 24mo of disbursement, free after", -0.2,
		[]FeeTier{{24, 3}, {0, 0}}},
	{"Axis LAP", "3%, but 25%/quarter part-pay free (not in Q1)", 0.1,
		[]FeeTier{{0, 3}}},
	{"AU Small Finance Bank", "5% within 12mo, 4% after, no part-pay allowed", 2.0,
		[]FeeTier{{12, 5}, {0, 4}}},
}

type LenderResult struct {
	Name         string
	Note         string
	RatePct      float64
	InterestPaid float64
	Fee          float64
	Total        float64
	IsBest       bool
	IsWorst      bool
}

func ComputeLeaderboard(P, rate float64, tenure, exit int, isFirm bool) []LenderResult {
	results := make([]LenderResult, 0, len(Lenders))
	for _, l := range Lenders {
		r2 := rate + l.DRate
		_, rows := Schedule(P, r2, tenure)
		bal := rows[exit-1].Balance

		var interestPaid float64
		for i := 0; i < exit; i++ {
			interestPaid += rows[i].Interest
		}

		feePct := l.FeeAt(exit)
		if !isFirm {
			feePct = 0
		}
		fee := bal * feePct / 100 * 1.18

		results = append(results, LenderResult{
			Name:         l.Name,
			Note:         l.Note,
			RatePct:      r2,
			InterestPaid: interestPaid,
			Fee:          fee,
			Total:        interestPaid + fee,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Total < results[j].Total })
	if len(results) > 0 {
		results[0].IsBest = true
		results[len(results)-1].IsWorst = true
	}
	return results
}
