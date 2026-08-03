package loancalc

import "testing"

func setup(idx, month int, lumpPct float64) (Lender, StrategyResult, LumpsumResult) {
	l := Lenders[idx]
	s := ComputeStrategy(3e7, 10.5, 60, month, true, idx)
	_, rows := Schedule(3e7, 10.5, 60)
	bal := rows[month-1].Balance
	lp := ComputeLumpsum(3e7, 10.5, 60, month, bal*lumpPct/100, idx)
	return l, s, lp
}

func TestRecommendPrefersLumpsumWhenBigger(t *testing.T) {
	l, s, lp := setup(0, 21, 25) // ICICI, big part-payment
	r := Recommend(l, s, lp)
	if !r.HasAction {
		t.Fatal("expected an actionable recommendation")
	}
	if r.Saving < lp.ReduceTenure.InterestSaved-1 {
		t.Errorf("saving %.0f should reflect the better of the two routes", r.Saving)
	}
}

func TestRecommendZeroFeeLenderSaysSo(t *testing.T) {
	l, s, lp := setup(6, 21, 0) // Canara, no lumpsum
	r := Recommend(l, s, lp)
	if r.HasAction {
		t.Error("zero-fee lender with no lumpsum should have no action")
	}
	if r.Headline != "Nothing to work around here" {
		t.Errorf("unexpected headline: %s", r.Headline)
	}
}

func TestRecommendAUWithNoLumpsumHasNoAction(t *testing.T) {
	l, s, lp := setup(5, 21, 0) // AU, no part-pay allowed, no lumpsum
	r := Recommend(l, s, lp)
	if r.HasAction {
		t.Error("AU with no lumpsum should have no actionable move")
	}
}

func TestRecommendAlwaysHasHeadlineAndDetail(t *testing.T) {
	for idx := range Lenders {
		for _, pct := range []float64{0, 25} {
			l, s, lp := setup(idx, 21, pct)
			r := Recommend(l, s, lp)
			if r.Headline == "" || r.Detail == "" {
				t.Errorf("lender %d pct %.0f: empty headline or detail", idx, pct)
			}
		}
	}
}
