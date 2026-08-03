package loancalc

import (
	"math"
	"testing"
)

// Figures below were verified independently before being asserted here.
func TestLumpsumBuilderCase(t *testing.T) {
	// 3 Cr, 10.5%, 60mo, at month 21, part-paying 25% of outstanding.
	// ICICI LAP is index 0 with DRate 0 so rate stays 10.5%.
	outstanding := 21228325.0
	lump := outstanding * 0.25

	r := ComputeLumpsum(3e7, 10.5, 60, 21, lump, 0)

	if math.Abs(r.Outstanding-outstanding) > 2000 {
		t.Errorf("Outstanding = %.0f, want approx %.0f", r.Outstanding, outstanding)
	}
	if math.Abs(r.Baseline.TotalInterest-3919538) > 5000 {
		t.Errorf("baseline interest = %.0f, want approx 3,919,538", r.Baseline.TotalInterest)
	}
	if math.Abs(r.ReduceEMI.InterestSaved-979885) > 5000 {
		t.Errorf("reduce-EMI saved = %.0f, want approx 979,885", r.ReduceEMI.InterestSaved)
	}
	if math.Abs(r.ReduceTenure.InterestSaved-1824821) > 5000 {
		t.Errorf("reduce-tenure saved = %.0f, want approx 1,824,821", r.ReduceTenure.InterestSaved)
	}
	if r.ReduceTenure.Months != 28 {
		t.Errorf("reduce-tenure months = %d, want 28", r.ReduceTenure.Months)
	}
	if math.Abs(r.ReduceEMI.EMIReduction-161204) > 2000 {
		t.Errorf("EMI reduction = %.0f, want approx 161,204", r.ReduceEMI.EMIReduction)
	}
}

func TestTenureRouteAlwaysBeatsEMIRoute(t *testing.T) {
	for _, month := range []int{6, 18, 30, 45} {
		r := ComputeLumpsum(3e7, 10.5, 60, month, 2000000, 0)
		if r.ReduceTenure.InterestSaved <= r.ReduceEMI.InterestSaved {
			t.Errorf("month %d: tenure route (%.0f) should beat EMI route (%.0f)",
				month, r.ReduceTenure.InterestSaved, r.ReduceEMI.InterestSaved)
		}
		if r.TenureAdvantage <= 0 {
			t.Errorf("month %d: TenureAdvantage = %.0f, want positive", month, r.TenureAdvantage)
		}
	}
}

func TestFreePartPayAllowanceByLender(t *testing.T) {
	cases := []struct {
		idx   int
		month int
		want  float64
		name  string
	}{
		{0, 18, 100, "ICICI LAP part-pay free"},
		{2, 18, 25, "HDFC annual 25% ladder"},
		{4, 18, 25, "Axis quarterly 25% ladder"},
		{5, 18, 0, "AU no part-pay allowed"},
		{6, 18, 100, "Canara zero fee"},
		{3, 12, 0, "SBI inside 24mo window"},
		{3, 30, 100, "SBI past 24mo window"},
	}
	for _, c := range cases {
		got := Lenders[c.idx].FreePartPayPct(c.month)
		if got != c.want {
			t.Errorf("%s: FreePartPayPct(%d) = %.0f, want %.0f", c.name, c.month, got, c.want)
		}
	}
}

func TestLumpsumClampedToOutstanding(t *testing.T) {
	r := ComputeLumpsum(3e7, 10.5, 60, 21, 9e9, 0)
	if r.Lumpsum > r.Outstanding+1 {
		t.Errorf("lumpsum %.0f should be clamped to outstanding %.0f", r.Lumpsum, r.Outstanding)
	}
	if r.NewBalance > 0.01 {
		t.Errorf("paying full outstanding should clear the loan, got balance %.2f", r.NewBalance)
	}
}

func TestLumpsumFreeFlag(t *testing.T) {
	// Axis (idx 4) allows 25% per quarter free
	r := ComputeLumpsum(3e7, 10.5, 60, 21, 1000000, 4)
	if !r.LumpsumIsFree {
		t.Error("1L on Axis should be within the free 25% allowance")
	}
	r2 := ComputeLumpsum(3e7, 10.5, 60, 21, 20000000, 4)
	if r2.LumpsumIsFree {
		t.Error("2 Cr on Axis should exceed the free 25% allowance")
	}
}
