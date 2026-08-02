package loancalc

import (
	"math"
	"testing"
)

func TestEMIMatchesKnownCase(t *testing.T) {
	// 3 Cr, 10.5%, 60 months -> should match the ~644,817 figure from the brief and JS POC
	got := EMI(3e7, 10.5, 60)
	want := 644817.0
	if math.Abs(got-want) > 500 {
		t.Errorf("EMI = %.2f, want approx %.2f", got, want)
	}
}

func TestScheduleOutstandingAtMonth18(t *testing.T) {
	_, rows := Schedule(3e7, 10.5, 60)
	got := rows[17].Balance // month 18, 0-indexed
	want := 2.26e7
	if math.Abs(got-want) > 5e5 {
		t.Errorf("balance at month 18 = %.2f, want approx %.2f", got, want)
	}
}
