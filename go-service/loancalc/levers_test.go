package loancalc

import "testing"

// The central fix: no lender should ever leave a running borrower with nothing.
func TestEveryLenderOffersAtLeastOneLever(t *testing.T) {
	for i, l := range Lenders {
		for _, tc := range []struct {
			P, rate      float64
			tenure, exit int
		}{
			{5e7, 9.5, 240, 120},
			{3e7, 10.5, 60, 18},
			{3e7, 10.5, 60, 40},
		} {
			p := BuildPlan(tc.P, tc.rate, tc.tenure, tc.exit, true, i)
			if !p.HasAction {
				t.Errorf("%s at month %d of %d: no lever offered", l.Name, tc.exit, tc.tenure)
			}
		}
	}
}

// Zero-fee lenders previously produced a single dead-end step.
func TestZeroFeeLendersStillGetInterestAdvice(t *testing.T) {
	for _, idx := range []int{6, 7} { // Canara, Indian Bank
		p := BuildPlan(5e7, 9.5, 240, 120, true, idx)
		if !p.HasAction {
			t.Fatalf("%s: expected an interest-reduction lever", Lenders[idx].Name)
		}
		if p.Best.Saving <= 0 {
			t.Errorf("%s: best lever saves nothing", Lenders[idx].Name)
		}
	}
}

// AU forbids part-payment, so it must still get the EMI lever.
func TestAUGetsEMILeverDespiteNoPartPay(t *testing.T) {
	p := BuildPlan(5e7, 9.5, 240, 120, true, 5)
	found := false
	for _, lv := range p.Levers {
		if lv.Kind == LeverEMIBoost {
			found = true
		}
		if lv.Kind == LeverPartPay {
			t.Error("AU does not permit part-payment, that lever must not appear")
		}
	}
	if !found {
		t.Error("AU should still get the EMI increase lever")
	}
}

func TestExitLeversRankedBySaving(t *testing.T) {
	for i := range Lenders {
		p := BuildPlan(3e7, 10.5, 60, 18, true, i)
		for j := 1; j < len(p.ExitLevers); j++ {
			if p.ExitLevers[j].Saving > p.ExitLevers[j-1].Saving {
				t.Errorf("%s: exit levers not ranked by fee avoided", Lenders[i].Name)
			}
		}
	}
}

func TestHoldLeversRankedByReturnPerRupee(t *testing.T) {
	for i := range Lenders {
		p := BuildPlan(5e7, 9.5, 240, 120, true, i)
		for j := 1; j < len(p.HoldLevers); j++ {
			if p.HoldLevers[j].PerRupee > p.HoldLevers[j-1].PerRupee {
				t.Errorf("%s: hold levers not ranked by return per rupee", Lenders[i].Name)
			}
		}
		for _, lv := range p.HoldLevers {
			if lv.Committed <= 0 {
				t.Errorf("%s: hold lever %q must state what it costs", Lenders[i].Name, lv.Title)
			}
		}
	}
}

func TestFeeLeversNeverRankedAgainstCashLevers(t *testing.T) {
	// A fee strategy costs nothing extra, so it must not be compared on the
	// same scale as a lever demanding lakhs in cash.
	p := BuildPlan(5e7, 9.5, 240, 120, true, 0) // ICICI, has both kinds
	for _, lv := range p.ExitLevers {
		if lv.Kind != LeverFee {
			t.Errorf("exit group must contain only fee levers, found %s", lv.Kind)
		}
	}
	for _, lv := range p.HoldLevers {
		if lv.Kind == LeverFee {
			t.Errorf("hold group must not contain fee levers")
		}
	}
}

func TestEveryLeverHasCompleteSteps(t *testing.T) {
	for i, l := range Lenders {
		p := BuildPlan(3e7, 10.5, 60, 18, true, i)
		for _, lv := range p.Levers {
			if len(lv.Steps) == 0 {
				t.Errorf("%s lever %s: no steps", l.Name, lv.Kind)
			}
			for k, st := range lv.Steps {
				if st.When == "" || st.Action == "" || st.Why == "" {
					t.Errorf("%s lever %s step %d: incomplete", l.Name, lv.Kind, k)
				}
			}
		}
	}
}
