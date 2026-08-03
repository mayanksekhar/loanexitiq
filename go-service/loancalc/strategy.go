package loancalc

import "math"

type DetailRow struct {
	Label string
	Value string
}

type StrategyResult struct {
	LenderName    string
	Type          StrategyType
	Title         string
	Mechanics     string
	Outstanding   float64
	FeePctNow     float64
	FeeNow        float64
	GstNow        float64
	TotalNow      float64
	HasStrategy   bool
	TotalStrategy float64
	SaveAmount    float64
	Rows          []DetailRow
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
	case StrategyNone, StrategyZeroFee:
		res.HasStrategy = false

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
		for p := startPeriod; p < s.Periods && balS > 1; p++ {
			for mo := 0; mo < periodMonths && balS > 1; mo++ {
				iMo := balS * r
				intCarried += iMo
				balS -= math.Min(E2-iMo, balS)
			}
			balS -= math.Min(lump, balS)
		}
		feeStag := 0.0
		if balS > 1 {
			feeStag = balS * feePctNow / 100 * 1.18
		}
		total := intCarried + feeStag

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
		}
	}

	return res
}
