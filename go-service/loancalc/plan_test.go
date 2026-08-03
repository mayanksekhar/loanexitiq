package loancalc

import (
	"math"
	"strings"
	"testing"
)

// Every lender must produce a plan with concrete, explained steps.
func TestEveryLenderProducesAPlan(t *testing.T) {
	for i, l := range Lenders {
		res := ComputeStrategy(3e7, 10.5, 60, 18, true, i)
		if len(res.Steps) == 0 {
			t.Errorf("%s: no plan steps produced", l.Name)
			continue
		}
		for j, st := range res.Steps {
			if st.When == "" || st.Action == "" || st.Why == "" {
				t.Errorf("%s step %d: incomplete (when=%q action=%q why=%q)", l.Name, j, st.When, st.Action, st.Why)
			}
			if !strings.Contains(st.Why, l.Name) && !strings.Contains(st.Why, "fee") && !strings.Contains(st.Why, "window") {
				t.Errorf("%s step %d: reason not grounded in a clause: %q", l.Name, j, st.Why)
			}
		}
	}
}

// The stub plan must clear the whole balance: part-payment plus final residue.
func TestStubPlanClearsBalance(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 0) // ICICI
	if len(res.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(res.Steps))
	}
	partPay := res.Steps[0].Amount
	residue := res.Steps[2].Amount
	stub := res.Outstanding - partPay
	// residue is the stub after EMIs during the wait, so it must be smaller
	if residue >= stub {
		t.Errorf("residue %.0f should be less than stub %.0f after paying EMIs", residue, stub)
	}
	if partPay <= 0 || partPay >= res.Outstanding {
		t.Errorf("part-payment %.0f out of range for outstanding %.0f", partPay, res.Outstanding)
	}
}

// The staggered ladder tranches must not exceed the free cap.
func TestStaggeredTranchesRespectCap(t *testing.T) {
	for _, idx := range []int{2, 4} { // HDFC annual, Axis quarterly
		res := ComputeStrategy(3e7, 10.5, 60, 18, true, idx)
		l := Lenders[idx]
		capAmt := res.Outstanding * l.Strategy.CapPct / 100
		for j, st := range res.Steps {
			if st.Amount > capAmt*1.01 {
				t.Errorf("%s step %d: tranche %.0f exceeds free cap %.0f", l.Name, j, st.Amount, capAmt)
			}
		}
	}
}

// A recommended payment must never exceed what is actually owed.
func TestRecommendedPayWithinOutstanding(t *testing.T) {
	for i, l := range Lenders {
		for _, month := range []int{6, 18, 40} {
			res := ComputeStrategy(3e7, 10.5, 60, month, true, i)
			if res.RecommendedPay > res.Outstanding+1 {
				t.Errorf("%s month %d: recommends %.0f but only %.0f is owed", l.Name, month, res.RecommendedPay, res.Outstanding)
			}
			if res.RecommendedPay < 0 {
				t.Errorf("%s month %d: negative recommendation %.0f", l.Name, month, res.RecommendedPay)
			}
		}
	}
}

// The seasoning plan must close exactly at the window boundary.
func TestSeasoningPlanWaitsExactlyToWindow(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 12, true, 3) // SBI, 24mo window, currently month 12
	if len(res.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(res.Steps))
	}
	if !strings.Contains(res.Steps[1].When, "24") {
		t.Errorf("plan should close at month 24, got %q", res.Steps[1].When)
	}
}

// Zero-fee lenders must not invent a strategy.
func TestZeroFeePlanIsSingleStep(t *testing.T) {
	for _, idx := range []int{6, 7} {
		res := ComputeStrategy(3e7, 10.5, 60, 18, true, idx)
		if len(res.Steps) != 1 {
			t.Errorf("%s: zero-fee lender should have exactly 1 step, got %d", Lenders[idx].Name, len(res.Steps))
		}
		if math.Abs(res.Steps[0].Amount-res.Outstanding) > 1 {
			t.Errorf("%s: should pay exactly the outstanding", Lenders[idx].Name)
		}
	}
}

func TestStaggeredPlanHasNoEmptyStepsAndCloses(t *testing.T) {
	for _, idx := range []int{2, 4} {
		for _, month := range []int{6, 12, 30} {
			res := ComputeStrategy(3e7, 10.5, 60, month, true, idx)
			for j, st := range res.Steps {
				if strings.Contains(st.Action, "Part-pay Rs 0") {
					t.Errorf("%s month %d step %d: meaningless zero part-payment", Lenders[idx].Name, month, j)
				}
			}
			last := res.Steps[len(res.Steps)-1].Action
			if !strings.Contains(last, "Close") && !strings.Contains(last, "fully cleared") {
				t.Errorf("%s month %d: plan does not end by closing the loan, ends with %q", Lenders[idx].Name, month, last)
			}
		}
	}
}
