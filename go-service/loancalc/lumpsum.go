package loancalc

// FreePartPayPct returns the share of outstanding principal this lender allows
// to be part-prepaid with no charge, at the given month. Derived from the
// documented exit clauses in Lenders. See docs/SOURCES.md.
func (l Lender) FreePartPayPct(month int) float64 {
	switch l.Strategy.Type {
	case StrategyStub:
		// Part-payments themselves are free; the fee only bites on full
		// foreclosure inside the clawback window.
		return 100
	case StrategyStaggered:
		return l.Strategy.CapPct
	case StrategyZeroFee:
		return 100
	case StrategySeasoning:
		if month >= l.Strategy.SeasoningMonths {
			return 100
		}
		return 0
	case StrategyNone:
		return 0
	}
	return 0
}

// LumpsumOption is one of the routes a borrower can take after part-paying.
type LumpsumOption struct {
	Months        int
	EMI           float64
	TotalInterest float64
	InterestSaved float64
	EMIReduction  float64
	MonthsSaved   int
}

type LumpsumResult struct {
	Outstanding     float64
	CurrentEMI      float64
	RemainingMonth  int
	Lumpsum         float64
	NewBalance      float64
	FreeAllowance   float64
	FreePct         float64
	LumpsumIsFree   bool
	Baseline        LumpsumOption
	ReduceEMI       LumpsumOption
	ReduceTenure    LumpsumOption
	TenureAdvantage float64
}

// clampZero removes sub-rupee floating point noise so differences that are
// mathematically zero print as zero rather than as a tiny negative.
func clampZero(v float64) float64 {
	if v > -0.5 && v < 0.5 {
		return 0
	}
	return v
}

// simulate pays a fixed instalment until the balance clears.
func simulate(bal, annualRate, instalment float64) (months int, totalInterest float64) {
	r := annualRate / 1200
	for bal > 0.01 && months < 10000 {
		i := bal * r
		if instalment <= i {
			// instalment cannot cover interest; loan would never amortize
			return months, totalInterest
		}
		totalInterest += i
		pay := instalment - i
		if pay > bal {
			pay = bal
		}
		bal -= pay
		months++
	}
	return months, totalInterest
}

// ComputeLumpsum models a part-payment at currentMonth and the two routes
// a borrower can take afterwards.
func ComputeLumpsum(P, rate float64, tenure, currentMonth int, lump float64, lenderIdx int) LumpsumResult {
	l := Lenders[lenderIdx]
	rate2 := rate

	e, rows := Schedule(P, rate2, tenure)
	bal := rows[currentMonth-1].Balance
	remaining := tenure - currentMonth

	if lump < 0 {
		lump = 0
	}
	if lump > bal {
		lump = bal
	}
	newBal := bal - lump

	freePct := l.FreePartPayPct(currentMonth)
	freeAllowance := bal * freePct / 100

	baseM, baseI := simulate(bal, rate2, e)

	res := LumpsumResult{
		Outstanding:    bal,
		CurrentEMI:     e,
		RemainingMonth: remaining,
		Lumpsum:        lump,
		NewBalance:     newBal,
		FreeAllowance:  freeAllowance,
		FreePct:        freePct,
		LumpsumIsFree:  lump <= freeAllowance+1,
		Baseline: LumpsumOption{
			Months:        baseM,
			EMI:           e,
			TotalInterest: baseI,
		},
	}

	if newBal <= 0.01 {
		res.ReduceEMI = LumpsumOption{Months: 0, EMI: 0, TotalInterest: 0, InterestSaved: baseI, EMIReduction: e, MonthsSaved: baseM}
		res.ReduceTenure = res.ReduceEMI
		return res
	}

	// Route A: keep the end date, shrink the instalment.
	newE := EMI(newBal, rate2, remaining)
	aM, aI := simulate(newBal, rate2, newE)
	res.ReduceEMI = LumpsumOption{
		Months:        aM,
		EMI:           newE,
		TotalInterest: aI,
		InterestSaved: clampZero(baseI - aI),
		EMIReduction:  clampZero(e - newE),
		MonthsSaved:   baseM - aM,
	}

	// Route B: keep the instalment, finish sooner.
	bM, bI := simulate(newBal, rate2, e)
	res.ReduceTenure = LumpsumOption{
		Months:        bM,
		EMI:           e,
		TotalInterest: bI,
		InterestSaved: clampZero(baseI - bI),
		EMIReduction:  0,
		MonthsSaved:   baseM - bM,
	}

	res.TenureAdvantage = clampZero(res.ReduceTenure.InterestSaved - res.ReduceEMI.InterestSaved)
	return res
}
