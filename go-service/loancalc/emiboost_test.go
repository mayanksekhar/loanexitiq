package loancalc

import (
	"math"
	"testing"
)

// Figures verified independently before assertion.
// 5 Cr at 9.5% over 240 months, at month 120: outstanding 3,60,18,113, EMI 4,66,065.
func TestEMIBoostSBICase(t *testing.T) {
	r := ComputeEMIBoost(5e7, 9.5, 240, 120)
	if !r.Applicable {
		t.Fatal("EMI boost should apply to any running loan")
	}
	if math.Abs(r.Outstanding-36018113) > 2000 {
		t.Errorf("outstanding = %.0f, want approx 3,60,18,113", r.Outstanding)
	}
	if math.Abs(r.Best.ExtraPerMonth-46607) > 500 {
		t.Errorf("10%% extra = %.0f, want approx 46,607", r.Best.ExtraPerMonth)
	}
	if math.Abs(r.Best.InterestSaved-3112597) > 20000 {
		t.Errorf("10%% saving = %.0f, want approx 31,12,597", r.Best.InterestSaved)
	}
	if r.Best.MonthsSaved != 16 {
		t.Errorf("months saved = %d, want 16", r.Best.MonthsSaved)
	}
}

func TestEMIBoostSavingsRiseWithIncrease(t *testing.T) {
	r := ComputeEMIBoost(5e7, 9.5, 240, 120)
	for i := 1; i < len(r.Options); i++ {
		if r.Options[i].InterestSaved <= r.Options[i-1].InterestSaved {
			t.Errorf("saving should rise with increase: %.0f%% saved %.0f, %.0f%% saved %.0f",
				r.Options[i-1].PctIncrease, r.Options[i-1].InterestSaved,
				r.Options[i].PctIncrease, r.Options[i].InterestSaved)
		}
		if r.Options[i].MonthsSaved < r.Options[i-1].MonthsSaved {
			t.Errorf("months saved should rise with increase")
		}
	}
}

// The lever must work regardless of lender, since it involves no fee at all.
func TestEMIBoostAppliesToZeroFeeLoans(t *testing.T) {
	r := ComputeEMIBoost(3e7, 10.5, 60, 18)
	if !r.Applicable {
		t.Error("EMI boost must apply even where there is no foreclosure fee")
	}
	if r.Best.InterestSaved <= 0 {
		t.Errorf("expected a positive saving, got %.0f", r.Best.InterestSaved)
	}
}

func TestEMIBoostNotApplicableAtVeryEndOfTenure(t *testing.T) {
	r := ComputeEMIBoost(3e7, 10.5, 60, 59)
	if r.Applicable {
		t.Error("should not claim a meaningful boost with one month left")
	}
}
