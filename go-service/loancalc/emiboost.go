package loancalc

// EMIBoostOption is one "pay a bigger instalment" scenario.
type EMIBoostOption struct {
	PctIncrease   float64
	NewEMI        float64
	ExtraPerMonth float64
	MonthsSaved   int
	TotalInterest float64
	InterestSaved float64
}

// EMIBoostResult holds the baseline and a ladder of increase options.
// This lever needs no lumpsum and no lender permission, so it applies to
// every loan including those with no foreclosure fee at all.
type EMIBoostResult struct {
	Outstanding      float64
	CurrentEMI       float64
	BaselineMonths   int
	BaselineInterest float64
	Options          []EMIBoostOption
	Best             EMIBoostOption
	Applicable       bool
}

// ComputeEMIBoost models raising the monthly instalment on the current balance.
func ComputeEMIBoost(P, rate float64, tenure, currentMonth int) EMIBoostResult {
	e, rows := Schedule(P, rate, tenure)
	bal := rows[currentMonth-1].Balance
	remaining := tenure - currentMonth

	res := EMIBoostResult{Outstanding: bal, CurrentEMI: e, BaselineMonths: remaining}
	if bal <= 1 || remaining <= 1 {
		return res
	}

	baseM, baseI := simulate(bal, rate, e)
	res.BaselineInterest = baseI

	for _, pct := range []float64{5, 10, 15, 20} {
		newE := e * (1 + pct/100)
		m, i := simulate(bal, rate, newE)
		if m <= 0 {
			continue
		}
		opt := EMIBoostOption{
			PctIncrease:   pct,
			NewEMI:        newE,
			ExtraPerMonth: newE - e,
			MonthsSaved:   baseM - m,
			TotalInterest: i,
			InterestSaved: clampZero(baseI - i),
		}
		res.Options = append(res.Options, opt)
		if opt.PctIncrease == 10 {
			res.Best = opt
		}
	}

	res.Applicable = len(res.Options) > 0 && res.Best.InterestSaved > 0
	return res
}
