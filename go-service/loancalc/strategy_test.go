package loancalc

import (
	"math"
	"testing"
)

func TestICICIStubStrategySavingMatchesBrief(t *testing.T) {
	// ICICI LAP is index 0 in StrategyLenders
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 0)

	if res.Type != StrategyStub {
		t.Fatalf("expected StrategyStub, got %s", res.Type)
	}
	if !res.HasStrategy {
		t.Fatal("expected HasStrategy to be true for ICICI LAP")
	}

	want := 872886.0 // matches the brief's ~Rs 8.7L net saving from the stub strategy
	if math.Abs(res.SaveAmount-want) > 5000 {
		t.Errorf("SaveAmount = %.2f, want approx %.2f", res.SaveAmount, want)
	}
}

func TestNoneStrategyHasNoSaving(t *testing.T) {
	// Kotak business is index 1, strategy type "none"
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 1)
	if res.HasStrategy {
		t.Error("expected HasStrategy to be false for a 'none' type lender")
	}
}

func TestStaggeredStrategyClearsToZeroResidual(t *testing.T) {
	// Axis staggered is index 4
	res := ComputeStrategy(3e7, 10.5, 60, 57, true, 4) // near tail of tenure, should be favorable
	if res.Type != StrategyStaggered {
		t.Fatalf("expected StrategyStaggered, got %s", res.Type)
	}
	if !res.HasStrategy {
		t.Fatal("expected HasStrategy to be true")
	}
}

func TestSeasoningStrategyAlreadyPastWaivesImmediately(t *testing.T) {
	// ICICI Instalment (MSME) is index 6, seasoning at 24 months
	res := ComputeStrategy(3e7, 10.5, 60, 30, true, 6) // exit month 30 > 24 month seasoning
	if res.TotalStrategy != 0 {
		t.Errorf("TotalStrategy = %.2f, want 0 (already past seasoning)", res.TotalStrategy)
	}
}
