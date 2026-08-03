package loancalc

// Recommendation is the single headline action for this borrower's situation.
type Recommendation struct {
	Headline   string
	Detail     string
	SavingText string
	Saving     float64
	HasAction  bool
}

// Recommend synthesizes the best available action from the exit strategy and
// the lumpsum routes, and says plainly when there is nothing worth doing.
func Recommend(l Lender, s StrategyResult, lp LumpsumResult) Recommendation {
	bestLump := lp.ReduceTenure
	lumpWorthIt := lp.Lumpsum > 0 && bestLump.InterestSaved > 0
	strategyWorthIt := s.HasStrategy && s.SaveAmount > 0

	switch {
	case strategyWorthIt && lumpWorthIt && s.SaveAmount >= bestLump.InterestSaved:
		return Recommendation{
			Headline:   "Use " + l.Name + "'s own clause to cut your exit fee",
			Detail:     s.Title + ". " + s.Mechanics,
			SavingText: "Saves " + formatINR(s.SaveAmount) + " against foreclosing today, more than the " + formatINR(bestLump.InterestSaved) + " a part-payment would save.",
			Saving:     s.SaveAmount,
			HasAction:  true,
		}

	case lumpWorthIt:
		detail := "Pay " + formatINR(lp.Lumpsum) + " against the principal and keep your EMI at " + formatINR(lp.CurrentEMI) +
			" instead of lowering it. The loan finishes " + itoa(bestLump.MonthsSaved) + " months sooner."
		if lp.LumpsumIsFree {
			detail += " " + l.Name + " allows this part-payment with no charge right now."
		}
		saving := "Saves " + formatINR(bestLump.InterestSaved) + " in interest."
		if lp.TenureAdvantage > 0 {
			saving += " Taking the lower EMI instead would save " + formatINR(lp.TenureAdvantage) + " less."
		}
		return Recommendation{
			Headline:   "Part-pay now, and keep your EMI the same",
			Detail:     detail,
			SavingText: saving,
			Saving:     bestLump.InterestSaved,
			HasAction:  true,
		}

	case strategyWorthIt:
		return Recommendation{
			Headline:   "Use " + l.Name + "'s own clause to cut your exit fee",
			Detail:     s.Title + ". " + s.Mechanics,
			SavingText: "Saves " + formatINR(s.SaveAmount) + " against foreclosing today.",
			Saving:     s.SaveAmount,
			HasAction:  true,
		}

	case s.Type == StrategyZeroFee:
		return Recommendation{
			Headline:  "Nothing to work around here",
			Detail:    l.Name + " charges no foreclosure or prepayment fee on this loan, so you can close it whenever you have the funds. The only cost of staying in the loan is the interest you keep paying.",
			HasAction: false,
		}

	default:
		return Recommendation{
			Headline:  "No fee-reducing move available on this loan",
			Detail:    "Neither " + l.Name + "'s published clauses nor a part-payment reduce your cost at this point. Closing the loan when you have the funds is the straightforward option. Always check your sanction letter for waivers that are not published.",
			HasAction: false,
		}
	}
}
