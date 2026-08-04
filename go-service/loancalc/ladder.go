package loancalc

// LadderResult models repeating part-payments capped per period, which is how
// Axis and HDFC actually allow free prepayment.
type LadderResult struct {
	Tranche       float64
	PeriodMonths  int
	Tranches      int
	TotalPaid     float64
	MonthsSaved   int
	InterestSaved float64
}

// ComputeLadder simulates paying the EMI plus a capped tranche every period.
func ComputeLadder(P, rate float64, tenure, exit int, capPct float64, periodMonths, maxTranches int, skipFirst bool) LadderResult {
	e, rows := Schedule(P, rate, tenure)
	bal := rows[exit-1].Balance
	r := rate / 1200

	baseM, baseI := simulate(bal, rate, e)

	tranche := bal * capPct / 100
	res := LadderResult{Tranche: tranche, PeriodMonths: periodMonths}

	cur := bal
	interest := 0.0
	months := 0
	period := 0
	if skipFirst {
		for i := 0; i < periodMonths && cur > 1; i++ {
			iM := cur * r
			interest += iM
			cur -= min2(e-iM, cur)
			months++
		}
	}
	for period < maxTranches && cur > 1 {
		for i := 0; i < periodMonths && cur > 1; i++ {
			iM := cur * r
			interest += iM
			cur -= min2(e-iM, cur)
			months++
		}
		pay := min2(tranche, cur)
		if pay < 1 {
			break
		}
		cur -= pay
		res.TotalPaid += pay
		res.Tranches++
		period++
	}
	// run out the remainder on normal EMI
	for cur > 1 && months < 10000 {
		iM := cur * r
		if e <= iM {
			break
		}
		interest += iM
		cur -= min2(e-iM, cur)
		months++
	}

	res.MonthsSaved = baseM - months
	res.InterestSaved = clampZero(baseI - interest)
	return res
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
