package loancalc

import "sort"

type Lender struct {
	Name  string
	Note  string
	DRate float64
	Fee   float64
}

var Lenders = []Lender{
	{"ICICI LAP", "4% non-individual, part-pay free", 0, 4},
	{"Kotak business", "4%, fixed rate product", 0, 4},
	{"HDFC", "~2.5% typical non-individual", -0.1, 2.5},
	{"SBI SME term", "~2% typical", -0.2, 2},
	{"Axis staggered", "25%/quarter part-pay = ~0%", 0.1, 0},
	{"Small Finance Bank", "12.5% rate, 3% fee, no part-pay", 2.0, 3},
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

		feePct := l.Fee
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
