package loancalc

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

type StrategyType string

const (
	StrategyStub      StrategyType = "stub"
	StrategyStaggered StrategyType = "staggered"
	StrategySeasoning StrategyType = "seasoning"
	StrategyNone      StrategyType = "none"
	StrategyZeroFee   StrategyType = "zerofee"
)

type Strategy struct {
	Type                StrategyType
	Title               string
	Mechanics           string
	ClawbackMonths      int
	CapPct              float64
	Periods             int
	PeriodMonths        int
	NoPrepayFirstPeriod bool
	SeasoningMonths     int
}

type Lender struct {
	Name     string
	Note     string
	FeeTiers []FeeTier
	Strategy Strategy
}

func (l Lender) FeeAt(exitMonth int) float64 {
	return feeAtTiers(l.FeeTiers, exitMonth)
}

// Unified lender list: cost data and exit-strategy data live together so one
// selection drives both. Verified against published fee schedules, Aug 2026.
// See docs/SOURCES.md.
var Lenders = []Lender{
	{
		Name: "ICICI LAP", Note: "4% non-individual, part-pay free, 12mo clawback",
		FeeTiers: []FeeTier{{0, 4}},
		Strategy: Strategy{
			Type: StrategyStub, Title: "Stub strategy: defuse the clawback",
			Mechanics:      "Part-payments are free, but full foreclosure within 12 months of a part-payment triggers a retroactive clawback of the fee on the part-paid amount. Part-pay almost everything today at Rs 0 charge, keep a small stub on EMI for 12 months to defuse the clawback, then foreclose only the residue.",
			ClawbackMonths: 12,
		},
	},
	{
		Name: "Kotak business", Note: "4% fixed rate, unsecured business loan",
		FeeTiers: []FeeTier{{0, 4}},
		Strategy: Strategy{
			Type: StrategyStub, Title: "Stub strategy: defuse the 12-month clawback",
			Mechanics:      "Kotak's business loan charges 4% plus GST on the outstanding at foreclosure, and an additional 4% plus GST on any amount part-prepaid in the preceding 12 months. Part-pay down to a small stub today, hold it on EMI for 12 months so that clawback window closes, then foreclose only the residue.",
			ClawbackMonths: 12,
		},
	},
	{
		Name: "HDFC LAP", Note: "up to 2.5% business use, 25%/yr part-pay free, MSME waiver",
		FeeTiers: []FeeTier{{0, 2.5}},
		Strategy: Strategy{
			Type: StrategyStaggered, Title: "Annual 25% part-pay ladder",
			Mechanics:    "One part-prepayment per year up to 25% of outstanding principal is free; the excess over that in the same year costs 2.5% plus GST. Spreading the payoff across several annual free tranches instead of one lump-sum foreclosure avoids most of the fee. MSME-classified borrowers closing from own funds get a full waiver instead, worth checking before using this ladder.",
			CapPct:       25,
			Periods:      4,
			PeriodMonths: 12,
		},
	},
	{
		Name: "SBI term loan", Note: "3% only within 24mo of disbursement, free after",
		FeeTiers: []FeeTier{{24, 3}, {0, 0}},
		Strategy: Strategy{
			Type: StrategySeasoning, Title: "Wait out the 24-month window",
			Mechanics:       "SBI only charges the 3% plus GST foreclosure fee if the loan is closed within 24 months of disbursement. After 24 months, foreclosure is free.",
			SeasoningMonths: 24,
		},
	},
	{
		Name: "Axis LAP", Note: "3%, but 25%/quarter part-pay free (not in Q1)",
		FeeTiers: []FeeTier{{0, 3}},
		Strategy: Strategy{
			Type: StrategyStaggered, Title: "Quarterly 25% part-pay ladder",
			Mechanics:           "No charge on part-prepayment up to 25% of principal outstanding per calendar quarter, with no prepayment allowed in the first quarter. Stagger the payoff across roughly five quarters instead of a lump-sum foreclosure.",
			CapPct:              25,
			Periods:             5,
			PeriodMonths:        3,
			NoPrepayFirstPeriod: true,
		},
	},
	{
		Name: "AU Small Finance Bank", Note: "5% within 12mo, 4% after, no part-pay allowed",
		FeeTiers: []FeeTier{{12, 5}, {0, 4}},
		Strategy: Strategy{
			Type: StrategyNone, Title: "No escape route: part-prepayment not allowed",
			Mechanics: "Part-prepayment is not permitted at all on this product. The tiered foreclosure fee, 5% within 12 months of last disbursement, 4% after, applies on the full outstanding with no legal lever to reduce it.",
		},
	},
	{
		Name: "Canara Bank LAP", Note: "nil prepayment and foreclosure charges (floating rate)",
		FeeTiers: []FeeTier{{0, 0}},
		Strategy: Strategy{
			Type: StrategyZeroFee, Title: "Already fee-free",
			Mechanics: "Canara Bank charges no prepayment or foreclosure fee at all on floating-rate loans against property. There's no fee to work around here, foreclosing today already costs nothing beyond the outstanding principal itself.",
		},
	},
	{
		Name: "Indian Bank LAP", Note: "nil prepayment and foreclosure charges (floating rate)",
		FeeTiers: []FeeTier{{0, 0}},
		Strategy: Strategy{
			Type: StrategyZeroFee, Title: "Already fee-free",
			Mechanics: "Indian Bank charges no prepayment or foreclosure fee at all on floating-rate loans against property. There's no fee to work around here, foreclosing today already costs nothing beyond the outstanding principal itself.",
		},
	},
	{
		Name: "ICICI Instalment (MSME)", Note: "different product, seasoning waiver",
		FeeTiers: []FeeTier{{0, 4}},
		Strategy: Strategy{
			Type: StrategySeasoning, Title: "MSME waiver or 24-month seasoning",
			Mechanics:       "A separate ICICI instalment product, not the LAP loan above. MSME classification plus closing with own funds waives the fee immediately; otherwise the fee drops to 0% after 24 months of seasoning.",
			SeasoningMonths: 24,
		},
	},
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

// ComputeForLender returns the full cost breakdown for a single lender by index.
func ComputeForLender(P, rate float64, tenure, exit int, isFirm bool, lenderIdx int) LenderResult {
	l := Lenders[lenderIdx]
	r2 := rate
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

	return LenderResult{
		Name:         l.Name,
		Note:         l.Note,
		RatePct:      r2,
		InterestPaid: interestPaid,
		Fee:          fee,
		FeePct:       feePct,
		Outstanding:  bal,
		Total:        interestPaid + fee,
	}
}
