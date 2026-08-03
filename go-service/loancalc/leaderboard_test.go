package loancalc

import (
	"math"
	"testing"
)

func TestICICITotalMatchesBrief(t *testing.T) {
	r := ComputeForLender(3e7, 10.5, 60, 18, true, 0)
	want := 5254344.0
	if math.Abs(r.Total-want) > 50000 {
		t.Errorf("ICICI LAP total = %.2f, want approx %.2f", r.Total, want)
	}
}

func TestSBIFeeTiering(t *testing.T) {
	sbi := Lenders[3]
	if sbi.Name != "SBI term loan" {
		t.Fatalf("index 3 is %q, expected SBI term loan", sbi.Name)
	}
	if got := sbi.FeeAt(18); got != 3 {
		t.Errorf("SBI fee at month 18 = %.1f, want 3", got)
	}
	if got := sbi.FeeAt(30); got != 0 {
		t.Errorf("SBI fee at month 30 = %.1f, want 0", got)
	}
}

func TestAUFeeTiering(t *testing.T) {
	au := Lenders[5]
	if au.Name != "AU Small Finance Bank" {
		t.Fatalf("index 5 is %q, expected AU Small Finance Bank", au.Name)
	}
	if got := au.FeeAt(10); got != 5 {
		t.Errorf("AU fee at month 10 = %.1f, want 5", got)
	}
	if got := au.FeeAt(18); got != 4 {
		t.Errorf("AU fee at month 18 = %.1f, want 4", got)
	}
}

func TestCanaraAndIndianAreZeroFee(t *testing.T) {
	for _, idx := range []int{6, 7} {
		l := Lenders[idx]
		if got := l.FeeAt(18); got != 0 {
			t.Errorf("%s fee = %.1f, want 0", l.Name, got)
		}
		r := ComputeForLender(3e7, 10.5, 60, 18, true, idx)
		if r.Fee != 0 {
			t.Errorf("%s computed fee = %.2f, want 0", l.Name, r.Fee)
		}
	}
}

func TestIndividualBorrowerPaysNoFee(t *testing.T) {
	r := ComputeForLender(3e7, 10.5, 60, 18, false, 0)
	if r.Fee != 0 {
		t.Errorf("individual borrower fee = %.2f, want 0", r.Fee)
	}
}
