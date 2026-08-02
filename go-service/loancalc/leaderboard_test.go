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
	want := 5254344.0 // matches the brief's Rs 52.5L total cost of exit at month 18
	if math.Abs(iciciTotal-want) > 50000 {
		t.Errorf("ICICI LAP total = %.2f, want approx %.2f", iciciTotal, want)
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
