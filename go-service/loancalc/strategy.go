package loancalc

import "math"

type StrategyType string

const (
	StrategyStub      StrategyType = "stub"
	StrategyStaggered StrategyType = "staggered"
	StrategySeasoning StrategyType = "seasoning"
	StrategyNone      StrategyType = "none"
)

type Strategy struct {
	Type            StrategyType
	Title           string
	Mechanics       string
	ClawbackMonths  int
	CapPct          float64
	Quarters        int
	SeasoningMonths int
	ImmediateFeePct float64
}

type StrategyLender struct {
	Name     string
	DRate    float64
	Strategy Strategy
}

var StrategyLenders = []StrategyLender{
	{"ICICI LAP", 0, Strategy{Type: StrategyStub, Title: "Stub strategy: defuse the clawback",
		Mechanics: "Part-payments are free, but full foreclosure within 12 months of a part-payment triggers a retroactive clawback of the fee on the part-paid amount. Part-pay almost everything today at Rs 0 charge, keep a small stub on EMI for 12 months to defuse the clawback, then foreclose only the residue.",
		ClawbackMonths: 12, ImmediateFeePct: 4}},
	{"Kotak business", 0, Strategy{Type: StrategyNone, Title: "No escape route: fixed rate",
		Mechanics: "Fixed-rate loan. The foreclosure fee applies regardless of RBI's Jan 2026 floating-rate prepayment ban, since that reform covers floating-rate loans only.",
		ImmediateFeePct: 4}},
	{"HDFC", -0.1, Strategy{Type: StrategyNone, Title: "No documented escape clause",
		Mechanics: "No published part-payment or seasoning waiver found for non-individual borrowers on this product.",
		ImmediateFeePct: 2.5}},
	{"SBI SME term", -0.2, Strategy{Type: StrategyNone, Title: "No documented escape clause",
		Mechanics: "No published part-payment or seasoning waiver found for this product.",
		ImmediateFeePct: 2}},
	{"Axis staggered", 0.1, Strategy{Type: StrategyStaggered, Title: "Quarterly 25% part-pay ladder",
		Mechanics: "No charge on part-prepayment up to 25% of principal outstanding per calendar quarter, with no prepayment allowed in the first quarter. Stagger the payoff across roughly five quarters instead of a lump-sum foreclosure.",
		CapPct: 25, Quarters: 5, ImmediateFeePct: 2}},
	{"Small Finance Bank", 2.0, Strategy{Type: StrategyNone, Title: "No escape route: worst structure",
		Mechanics: "Part-prepayment is not allowed at all on this product.",
		ImmediateFeePct: 3}},
	{"ICICI Instalment (MSME)", 0.2, Strategy{Type: StrategySeasoning, Title: "MSME waiver or 24-month seasoning",
		Mechanics: "A separate ICICI instalment product. MSME classification plus closing with own funds waives the fee immediately; otherwise the fee drops to 0% after 24 months of seasoning.",
		SeasoningMonths: 24, ImmediateFeePct: 4}},
}

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
	l := StrategyLenders[lenderIdx]
	s := l.Strategy
	rate2 := rate + l.DRate
	r := rate2 / 1200
	_, rows := Schedule(P, rate2, tenure)
	bal := rows[exit-1].Balance

	feePctNow := s.ImmediateFeePct
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
	case StrategyNone:
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
		for mo := 0; mo < 3 && balS > 1; mo++ {
			iMo := balS * r
			intCarried += iMo
			balS -= math.Min(E2-iMo, balS)
		}
		lump := balS * s.CapPct / 100
		for q := 1; q < s.Quarters && balS > 1; q++ {
			for mo := 0; mo < 3 && balS > 1; mo++ {
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
