package loancalc

import (
	"math"
	"testing"
)

func TestLeaderboardICICITotalMatchesBrief(t *testing.T) {
	results := ComputeLeaderboard(3e7, 10.5, 60, 18, true)
	var iciciTotal float64
	found := false
	for _, r := range results {
		if r.Name == "ICICI LAP" {
			iciciTotal = r.Total
			found = true
		}
	}
	if !found {
		t.Fatal("ICICI LAP not found in results")
	}
	want := 5254344.0
	if math.Abs(iciciTotal-want) > 50000 {
		t.Errorf("ICICI LAP total = %.2f, want approx %.2f", iciciTotal, want)
	}
}

func TestSBIFeeTiering(t *testing.T) {
	var sbi Lender
	for _, l := range Lenders {
		if l.Name == "SBI term loan" {
			sbi = l
		}
	}
	if got := sbi.FeeAt(18); got != 3 {
		t.Errorf("SBI fee at month 18 = %.1f, want 3 (within 24mo window)", got)
	}
	if got := sbi.FeeAt(30); got != 0 {
		t.Errorf("SBI fee at month 30 = %.1f, want 0 (past 24mo window)", got)
	}
}

func TestAUFeeTiering(t *testing.T) {
	var au Lender
	for _, l := range Lenders {
		if l.Name == "AU Small Finance Bank" {
			au = l
		}
	}
	if got := au.FeeAt(10); got != 5 {
		t.Errorf("AU fee at month 10 = %.1f, want 5", got)
	}
	if got := au.FeeAt(18); got != 4 {
		t.Errorf("AU fee at month 18 = %.1f, want 4", got)
	}
}

func TestLeaderboardSortedAscendingWithFlags(t *testing.T) {
	results := ComputeLeaderboard(3e7, 10.5, 60, 18, true)
	for i := 1; i < len(results); i++ {
		if results[i].Total < results[i-1].Total {
			t.Errorf("results not sorted ascending at index %d", i)
		}
	}
	if !results[0].IsBest {
		t.Error("first result should be marked IsBest")
	}
	if !results[len(results)-1].IsWorst {
		t.Error("last result should be marked IsWorst")
	}
}
