package loancalc

import "math"

type DetailRow struct {
	Label string
	Value string
}

// PlanStep is one concrete instruction in the exit plan.
type PlanStep struct {
	When   string
	Action string
	Amount float64
	Why    string
}

type StrategyResult struct {
	Steps          []PlanStep
	RecommendedPay float64
	LenderName     string
	Type           StrategyType
	Title          string
	Mechanics      string
	Outstanding    float64
	FeePctNow      float64
	FeeNow         float64
	GstNow         float64
	TotalNow       float64
	HasStrategy    bool
	TotalStrategy  float64
	SaveAmount     float64
	Rows           []DetailRow
}

func ComputeStrategy(P, rate float64, tenure, exit int, isFirm bool, lenderIdx int) StrategyResult {
	l := Lenders[lenderIdx]
	s := l.Strategy
	rate2 := rate
	r := rate2 / 1200
	_, rows := Schedule(P, rate2, tenure)
	bal := rows[exit-1].Balance

	feePctNow := l.FeeAt(exit)
	if !isFirm {
		feePctNow = 0
	}
	feeNow := bal * feePctNow / 100
	gstNow := feeNow * 0.18
	totalNow := feeNow + gstNow

	res := StrategyResult{
		LenderName:  l.Name,
		Type:        s.Type,
		Title:       s.Title,
		Mechanics:   s.Mechanics,
		Outstanding: bal,
		FeePctNow:   feePctNow,
		FeeNow:      feeNow,
		GstNow:      gstNow,
		TotalNow:    totalNow,
	}

	switch s.Type {
	case StrategyZeroFee:
		res.HasStrategy = false
		res.Steps = []PlanStep{{
			When: "Any time", Amount: bal,
			Action: "Close the loan by paying " + formatINR(bal),
			Why:    l.Name + " charges no foreclosure or prepayment fee, so there is nothing to time or work around.",
		}}

	case StrategyNone:
		res.HasStrategy = false
		res.Steps = []PlanStep{{
			When: "Today", Amount: bal + feeNow + gstNow,
			Action: "Close the loan by paying " + formatINR(bal+feeNow+gstNow),
			Why:    l.Name + " does not permit part-prepayment, so the fee of " + formatINR(totalNow) + " cannot be reduced by timing or staging.",
		}}

	case StrategyStub:
		nRem := tenure - exit
		stub := math.Min(1500000, bal*0.08)
		part := bal - stub
		eStub := stub * r * math.Pow(1+r, float64(nRem)) / (math.Pow(1+r, float64(nRem)) - 1)
		m := s.ClawbackMonths
		if nRem < m {
			m = nRem
		}
		bal12 := stub*math.Pow(1+r, float64(m)) - eStub*(math.Pow(1+r, float64(m))-1)/r
		feeB := math.Max(bal12, 0) * feePctNow / 100 * 1.18
		intStub := eStub*float64(m) - (stub - math.Max(bal12, 0))
		total := feeB + intStub

		res.HasStrategy = true
		res.TotalStrategy = total
		res.SaveAmount = totalNow - total
		res.Rows = []DetailRow{
			{"Part-pay today (free)", formatINR(part)},
			{"Stub kept on EMI", formatINR(stub)},
			{"Interest carried on stub", formatINR(intStub)},
			{"Fee on residue + GST", formatINR(feeB)},
		}
		res.RecommendedPay = part
		res.Steps = []PlanStep{
			{
				When: "Today", Amount: part,
				Action: "Part-pay " + formatINR(part),
				Why:    "Part-payments carry no charge at " + l.Name + ". This clears most of the balance for free.",
			},
			{
				When: "Next " + itoa(m) + " months", Amount: eStub,
				Action: "Keep paying EMI of " + formatINR(eStub) + " on the remaining " + formatINR(stub),
				Why:    "Foreclosing within " + itoa(s.ClawbackMonths) + " months of a part-payment claws back the fee on the amount you just part-paid. Waiting out this window defuses that.",
			},
			{
				When: "Month " + itoa(m), Amount: math.Max(bal12, 0),
				Action: "Close the remaining " + formatINR(math.Max(bal12, 0)),
				Why:    "The fee now applies only to this small residue, costing " + formatINR(feeB) + " instead of " + formatINR(totalNow) + ".",
			},
		}

	case StrategyStaggered:
		E2 := EMI(P, rate2, tenure)
		balS := bal
		intCarried := 0.0
		periodMonths := s.PeriodMonths
		if periodMonths == 0 {
			periodMonths = 3
		}
		startPeriod := 0
		if s.NoPrepayFirstPeriod {
			for mo := 0; mo < periodMonths && balS > 1; mo++ {
				iMo := balS * r
				intCarried += iMo
				balS -= math.Min(E2-iMo, balS)
			}
			startPeriod = 1
		}
		lump := balS * s.CapPct / 100
		res.RecommendedPay = lump
		if s.NoPrepayFirstPeriod {
			res.Steps = append(res.Steps, PlanStep{
				When: "Months 1 to " + itoa(periodMonths), Amount: 0,
				Action: "Pay your normal EMI only",
				Why:    l.Name + " does not allow part-prepayment in the first period, so nothing extra can be paid yet.",
			})
		}
		for p := startPeriod; p < s.Periods && balS > 1; p++ {
			for mo := 0; mo < periodMonths && balS > 1; mo++ {
				iMo := balS * r
				intCarried += iMo
				balS -= math.Min(E2-iMo, balS)
			}
			pay := math.Min(lump, balS)
			balS -= pay
			if pay < 1 {
				break
			}
			res.Steps = append(res.Steps, PlanStep{
				When: "Month " + itoa((p+1)*periodMonths), Amount: pay,
				Action: "Part-pay " + formatINR(pay),
				Why:    "Up to " + ratePct(s.CapPct) + " of outstanding per period is free at " + l.Name + ". Staying under that cap avoids the fee entirely.",
			})
		}
		feeStag := 0.0
		if balS > 1 {
			feeStag = balS * feePctNow / 100 * 1.18
		}
		total := intCarried + feeStag

		closeWhen := "After the ladder"
		if n := len(res.Steps); n > 0 {
			closeWhen = res.Steps[n-1].When
		}
		if balS > 1 {
			res.Steps = append(res.Steps, PlanStep{
				When: closeWhen, Amount: balS,
				Action: "Close the remaining " + formatINR(balS),
				Why:    "The fee now applies only to this residue, costing " + formatINR(feeStag) + " instead of " + formatINR(totalNow) + " if you had closed in one lump sum today.",
			})
		} else {
			res.Steps = append(res.Steps, PlanStep{
				When: closeWhen, Amount: 0,
				Action: "The loan is fully cleared",
				Why:    "The free tranches above pay off the entire balance, so no foreclosure fee is ever triggered. You avoid the " + formatINR(totalNow) + " a lump-sum exit would have cost.",
			})
		}

		res.HasStrategy = true
		res.TotalStrategy = total
		res.SaveAmount = totalNow - total
		res.Rows = []DetailRow{
			{"Interest carried across ladder", formatINR(intCarried)},
			{"Residual balance after ladder", formatINR(math.Max(balS, 0))},
			{"Fee on residue + GST", formatINR(feeStag)},
		}

	case StrategySeasoning:
		if exit >= s.SeasoningMonths {
			res.HasStrategy = true
			res.TotalStrategy = 0
			res.SaveAmount = totalNow
			res.Rows = []DetailRow{{"Loan age", "already past seasoning, fee waived"}}
			res.Steps = []PlanStep{{
				When: "Today", Amount: bal,
				Action: "Close the loan by paying " + formatINR(bal),
				Why:    "You are already past " + l.Name + "'s " + itoa(s.SeasoningMonths) + " month window, so there is no foreclosure fee to avoid.",
			}}
		} else {
			waitMonths := s.SeasoningMonths - exit
			nRem := tenure - exit
			wm := waitMonths
			if nRem < wm {
				wm = nRem
			}
			E2 := EMI(P, rate2, tenure)
			balAfter := bal*math.Pow(1+r, float64(wm)) - E2*(math.Pow(1+r, float64(wm))-1)/r
			intWait := E2*float64(wm) - (bal - math.Max(balAfter, 0))

			res.HasStrategy = true
			res.TotalStrategy = intWait
			res.SaveAmount = totalNow - intWait
			res.Rows = []DetailRow{
				{"Months to wait for seasoning", itoa(wm)},
				{"Interest carried while waiting", formatINR(intWait)},
			}
			res.Steps = []PlanStep{
				{
					When: "Next " + itoa(wm) + " months", Amount: E2,
					Action: "Keep paying your EMI of " + formatINR(E2),
					Why:    l.Name + " charges the fee only inside the first " + itoa(s.SeasoningMonths) + " months. You are month " + itoa(exit) + ", so " + itoa(wm) + " more months clears the window.",
				},
				{
					When: "Month " + itoa(exit+wm), Amount: math.Max(balAfter, 0),
					Action: "Close the remaining " + formatINR(math.Max(balAfter, 0)) + " with no fee",
					Why:    "Waiting costs " + formatINR(intWait) + " in interest, against a fee of " + formatINR(totalNow) + " if you closed today.",
				},
			}
		}
	}

	return res
}
