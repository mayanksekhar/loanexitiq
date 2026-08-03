package loancalc

import "testing"

func TestBuildChartBarCount(t *testing.T) {
	c := BuildChart(3e7, 10.5, 60, 18)
	if len(c.Bars) != 60 {
		t.Errorf("bar count = %d, want 60", len(c.Bars))
	}
}

func TestBuildChartDimsAfterExit(t *testing.T) {
	c := BuildChart(3e7, 10.5, 60, 18)
	if c.Bars[17].Dimmed {
		t.Error("bar at index 17 (month 18) should not be dimmed")
	}
	if !c.Bars[18].Dimmed {
		t.Error("bar at index 18 (month 19) should be dimmed")
	}
}

func TestBuildChartInterestShrinksOverTime(t *testing.T) {
	c := BuildChart(3e7, 10.5, 60, 60)
	if c.Bars[59].InterestH >= c.Bars[0].InterestH {
		t.Errorf("interest should shrink: first=%.2f last=%.2f", c.Bars[0].InterestH, c.Bars[59].InterestH)
	}
}

func TestBuildChartPrincipalGrowsOverTime(t *testing.T) {
	c := BuildChart(3e7, 10.5, 60, 60)
	if c.Bars[59].PrincipalH <= c.Bars[0].PrincipalH {
		t.Errorf("principal should grow: first=%.2f last=%.2f", c.Bars[0].PrincipalH, c.Bars[59].PrincipalH)
	}
}

func TestBuildChartInterestPctMatchesEarlyTenure(t *testing.T) {
	// At month 18 of 60 at 10.5%, most of each EMI is still interest.
	c := BuildChart(3e7, 10.5, 60, 18)
	if c.IntPctAtExit < 30 || c.IntPctAtExit > 45 {
		t.Errorf("interest pct at exit = %d, expected roughly 30-45", c.IntPctAtExit)
	}
}

func TestBuildChartHandlesShortTenure(t *testing.T) {
	c := BuildChart(5e6, 10.5, 24, 3)
	if len(c.Bars) != 24 {
		t.Errorf("bar count = %d, want 24", len(c.Bars))
	}
	if c.Bars[0].Width < 1 {
		t.Errorf("bar width = %.2f, should be at least 1", c.Bars[0].Width)
	}
}

func TestBuildChartPrincipalFiguresAtMonth21(t *testing.T) {
	c := BuildChart(3e7, 10.5, 60, 21)
	// Verified independently: paid 1.354 Cr, interest 47.7 L, principal 87.7 L, outstanding 2.12 Cr
	if c.TotalPaid < 1.34e7 || c.TotalPaid > 1.37e7 {
		t.Errorf("TotalPaid = %.0f, want approx 1.354 Cr", c.TotalPaid)
	}
	if c.PrincipalCleared < 8.6e6 || c.PrincipalCleared > 8.9e6 {
		t.Errorf("PrincipalCleared = %.0f, want approx 87.7 L", c.PrincipalCleared)
	}
	if c.Outstanding < 2.10e7 || c.Outstanding > 2.14e7 {
		t.Errorf("Outstanding = %.0f, want approx 2.12 Cr", c.Outstanding)
	}
	if c.IntPctAtExit != 35 {
		t.Errorf("IntPctAtExit = %d, want 35", c.IntPctAtExit)
	}
}

func TestBuildChartLongTenureShowsHighInterestShare(t *testing.T) {
	// 20-year home loan: the interest share should be dramatically higher, honestly
	c := BuildChart(5e6, 8.5, 240, 21)
	if c.IntPctAtExit < 75 {
		t.Errorf("IntPctAtExit for 20yr loan = %d, expected 75+", c.IntPctAtExit)
	}
}
