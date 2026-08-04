package loancalc

import "sort"

// LeverKind identifies which mechanism a lever uses.
type LeverKind string

const (
	LeverFee      LeverKind = "fee"      // avoid or reduce the foreclosure charge
	LeverPartPay  LeverKind = "partpay"  // pay a lumpsum against principal
	LeverEMIBoost LeverKind = "emiboost" // raise the monthly instalment
)

// Lever is one available course of action, with what it saves and how to do it.
type Lever struct {
	Kind      LeverKind
	Title     string
	Summary   string
	Requires  string
	Committed float64
	PerRupee  float64
	Saving    float64
	Steps     []PlanStep
}

// PlanResult is the single source of truth for both the headline recommendation
// and the detailed plan, so the two can never disagree.
type PlanResult struct {
	// ExitLevers apply when the borrower intends to close the loan. Their
	// saving is a charge avoided on money being paid anyway.
	ExitLevers []Lever
	// HoldLevers apply when the borrower keeps the loan running. Their saving
	// is interest avoided, and each demands real cash, so they rank on return
	// per rupee committed rather than on headline saving.
	HoldLevers  []Lever
	Levers      []Lever
	Best        Lever
	HasAction   bool
	NoActionWhy string

	// ExitCostToday is what closing right now costs in fee and GST.
	ExitCostToday float64
	// ExitVerdict explains the cheapest way out when no clause improves on
	// simply closing, including which strategies were modelled and rejected.
	ExitVerdict  string
	RejectedNote string
	CloseSteps   []PlanStep
}

// BuildPlan gathers every lever available on this loan and ranks them by saving.
func BuildPlan(P, rate float64, tenure, exit int, isFirm bool, lenderIdx int) PlanResult {
	l := Lenders[lenderIdx]
	var levers []Lever

	// 1. Fee lever: only exists where a fee exists and a clause can reduce it.
	s := ComputeStrategy(P, rate, tenure, exit, isFirm, lenderIdx)
	rejected := ""
	if s.HasStrategy && s.SaveAmount <= 0 {
		rejected = "We modelled " + lowerFirst(s.Title) + " for this loan. On your numbers it would cost " +
			formatINR(-s.SaveAmount) + " more than simply closing today, because the interest carried while waiting " +
			"outweighs the fee it avoids. We are not recommending it."
	}
	if s.HasStrategy && s.SaveAmount > 0 && len(s.Steps) > 0 {
		levers = append(levers, Lever{
			Kind:      LeverFee,
			Title:     s.Title,
			Requires:  "Closing the loan on the schedule below",
			Committed: 0,
			Summary:   "Cuts the exit charge by " + formatINR(s.SaveAmount) + " using " + l.Name + "'s own clause.",
			Saving:    s.SaveAmount,
			Steps:     s.Steps,
		})
	}

	// 2. Part-payment lever, shaped by how the lender actually caps prepayment.
	freePct := l.FreePartPayPct(exit)
	if freePct > 0 {
		_, rows := Schedule(P, rate, tenure)
		bal := rows[exit-1].Balance
		st := l.Strategy

		if st.Type == StrategyStaggered && st.PeriodMonths > 0 {
			// Capped per period, so the only lawful route is a repeating ladder.
			maxT := st.Periods
			if maxT < 1 {
				maxT = 4
			}
			ld := ComputeLadder(P, rate, tenure, exit, st.CapPct, st.PeriodMonths, maxT, st.NoPrepayFirstPeriod)
			// A ladder that commits most of the balance is repayment, not strategy.
			for ld.TotalPaid > bal*0.5 && maxT > 1 {
				maxT--
				ld = ComputeLadder(P, rate, tenure, exit, st.CapPct, st.PeriodMonths, maxT, st.NoPrepayFirstPeriod)
			}
			if ld.InterestSaved > 0 && ld.Tranches > 0 {
				var steps []PlanStep
				if st.NoPrepayFirstPeriod {
					steps = append(steps, PlanStep{
						When: "Months 1 to " + itoa(st.PeriodMonths), Amount: 0,
						Action: "Pay your normal EMI only",
						Why:    l.Name + " does not permit part-prepayment in the first period, so nothing extra can go in yet.",
					})
				}
				for k := 0; k < ld.Tranches; k++ {
					at := (k + 1) * ld.PeriodMonths
					if st.NoPrepayFirstPeriod {
						at += st.PeriodMonths
					}
					steps = append(steps, PlanStep{
						When: "Month " + itoa(at), Amount: ld.Tranche,
						Action: "Part-pay " + formatINR(ld.Tranche),
						Why:    l.Name + " allows " + ratePct(st.CapPct) + " of outstanding free every " + itoa(st.PeriodMonths) + " months. Staying at the cap avoids any charge.",
					})
				}
				levers = append(levers, Lever{
					Kind:      LeverPartPay,
					Title:     "Free part-payment ladder, " + ratePct(st.CapPct) + " every " + itoa(st.PeriodMonths) + " months",
					Summary:   "Ends the loan " + itoa(ld.MonthsSaved) + " months early using " + itoa(ld.Tranches) + " capped tranches, all charge free.",
					Requires:  formatINR(ld.Tranche) + " per tranche, " + formatINR(ld.TotalPaid) + " in total",
					Committed: ld.TotalPaid,
					Saving:    ld.InterestSaved,
					Steps:     steps,
				})
			}
		} else {
			// No per-period cap, so a single part-payment is permitted.
			lump := bal * 0.25
			lp := ComputeLumpsum(P, rate, tenure, exit, lump, lenderIdx)
			if lp.ReduceTenure.InterestSaved > 0 {
				cap := "with no cap on how much you may part-pay"
				if freePct < 100 {
					cap = "up to " + formatINR(lp.FreeAllowance)
				}
				steps := []PlanStep{
					{
						When: "Today", Amount: lp.Lumpsum,
						Action: "Part-pay " + formatINR(lp.Lumpsum) + " against the principal",
						Why:    l.Name + " permits part-payment " + cap + " at no charge, so this costs nothing beyond the money itself." + clawbackNote(l),
					},
					{
						When: "From next month", Amount: lp.CurrentEMI,
						Action: "Keep your EMI at " + formatINR(lp.CurrentEMI) + ". Do not let the bank lower it",
						Why:    "The bank will offer to reduce the instalment instead. Holding it steady ends the loan " + itoa(lp.ReduceTenure.MonthsSaved) + " months early and saves " + formatINR(lp.TenureAdvantage) + " more than the lower EMI would.",
					},
				}
				levers = append(levers, Lever{
					Kind:      LeverPartPay,
					Title:     "Part-pay now, and refuse the lower EMI",
					Summary:   "Ends the loan " + itoa(lp.ReduceTenure.MonthsSaved) + " months early.",
					Requires:  formatINR(lp.Lumpsum) + " available today",
					Committed: lp.Lumpsum,
					Saving:    lp.ReduceTenure.InterestSaved,
					Steps:     steps,
				})
			}
		}
	}

	// 3. EMI increase lever: needs no lumpsum and no lender permission, so it is
	// available on every running loan, including fee-free ones.
	eb := ComputeEMIBoost(P, rate, tenure, exit)
	if eb.Applicable {
		steps := []PlanStep{
			{
				When: "From next month", Amount: eb.Best.NewEMI,
				Action: "Raise your EMI from " + formatINR(eb.CurrentEMI) + " to " + formatINR(eb.Best.NewEMI),
				Why:    "That is " + formatINR(eb.Best.ExtraPerMonth) + " more each month, and every rupee of it goes straight to principal rather than interest.",
			},
			{
				When: "Result", Amount: 0,
				Action: "Loan ends " + itoa(eb.Best.MonthsSaved) + " months early, saving " + formatINR(eb.Best.InterestSaved),
				Why:    "No lender permission is needed to increase an instalment, and no foreclosure fee is involved, so this works on any loan.",
			},
		}
		levers = append(levers, Lever{
			Kind:      LeverEMIBoost,
			Title:     "Raise your EMI by 10 percent",
			Summary:   "Ends the loan " + itoa(eb.Best.MonthsSaved) + " months early. No lumpsum and no lender permission needed.",
			Requires:  formatINR(eb.Best.ExtraPerMonth) + " more each month",
			Committed: eb.Best.ExtraPerMonth * float64(eb.BaselineMonths-eb.Best.MonthsSaved),
			Saving:    eb.Best.InterestSaved,
			Steps:     steps,
		})
	}

	res := PlanResult{}
	for _, lv := range levers {
		if lv.Committed > 0 {
			lv.PerRupee = lv.Saving / lv.Committed
		}
		if lv.Kind == LeverFee {
			res.ExitLevers = append(res.ExitLevers, lv)
		} else {
			res.HoldLevers = append(res.HoldLevers, lv)
		}
	}
	// Exit levers rank on fee avoided; hold levers on return per rupee, since a
	// bigger headline saving that demands far more cash is not automatically
	// the better move.
	sort.SliceStable(res.ExitLevers, func(i, j int) bool { return res.ExitLevers[i].Saving > res.ExitLevers[j].Saving })
	sort.SliceStable(res.HoldLevers, func(i, j int) bool { return res.HoldLevers[i].PerRupee > res.HoldLevers[j].PerRupee })

	res.RejectedNote = rejected
	res.ExitCostToday = s.TotalNow
	res.CloseSteps = []PlanStep{{
		When: "Today", Amount: s.Outstanding + s.TotalNow,
		Action: "Pay " + formatINR(s.Outstanding+s.TotalNow) + " and the loan is closed",
		Why: formatINR(s.Outstanding) + " is the principal you still owe. " +
			closeFeeClause(s.TotalNow, l.Name),
	}}
	if len(res.ExitLevers) == 0 {
		switch {
		case s.TotalNow <= 0:
			res.ExitVerdict = l.Name + " charges no foreclosure fee at this point, so closing today is already the cheapest possible exit. There is nothing to time or work around."
		case rejected != "":
			res.ExitVerdict = "Closing today is your cheapest way out of this loan."
		default:
			res.ExitVerdict = "No clause in " + l.Name + "'s published schedule reduces the exit charge at this point, so closing today is the cheapest way out. Check your sanction letter for waivers that are not published."
		}
	}

	res.Levers = append(append([]Lever{}, res.ExitLevers...), res.HoldLevers...)
	if len(res.ExitLevers) > 0 {
		res.Best = res.ExitLevers[0]
		res.HasAction = true
	} else if len(res.HoldLevers) > 0 {
		res.Best = res.HoldLevers[0]
		res.HasAction = true
	} else {
		res.NoActionWhy = "This loan is close enough to the end of its term that no part-payment, instalment increase or clause in " + l.Name + "'s schedule would meaningfully reduce what you pay. Closing it when you have the funds is the straightforward option."
	}
	return res
}

// clawbackNote warns where part-paying then closing soon would trigger a clawback.
func clawbackNote(l Lender) string {
	if l.Strategy.Type == StrategyStub && l.Strategy.ClawbackMonths > 0 {
		return " Important: if you then close the loan entirely within " + itoa(l.Strategy.ClawbackMonths) +
			" months of this part-payment, " + l.Name + " claws the fee back on the amount you just paid. Either keep the loan running past that window, or use the clause strategy listed separately."
	}
	return ""
}

func closeFeeClause(fee float64, lender string) string {
	if fee <= 0 {
		return lender + " adds no foreclosure fee, so that is the whole amount."
	}
	return "On top of that, " + lender + " charges " + formatINR(fee) + " in foreclosure fee and GST."
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	b := []rune(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] = b[0] + 32
	}
	return string(b)
}
