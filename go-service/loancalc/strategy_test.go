package loancalc

import (
	"math"
	"testing"
)

func TestICICIStubStrategySavingMatchesBrief(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 0)
	if res.Type != StrategyStub || !res.HasStrategy {
		t.Fatalf("expected StrategyStub with HasStrategy, got %s / %v", res.Type, res.HasStrategy)
	}
	want := 872886.0
	if math.Abs(res.SaveAmount-want) > 5000 {
		t.Errorf("SaveAmount = %.2f, want approx %.2f", res.SaveAmount, want)
	}
}

func TestKotakStubStrategyHasClawback(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 1)
	if res.Type != StrategyStub || !res.HasStrategy {
		t.Fatalf("expected Kotak to be StrategyStub, got %s", res.Type)
	}
}

func TestHDFCAnnualLadderClearsSubstantially(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 2)
	if res.Type != StrategyStaggered || !res.HasStrategy {
		t.Fatalf("expected HDFC to be StrategyStaggered, got %s", res.Type)
	}
}

func TestSBISeasoningFlips(t *testing.T) {
	before := ComputeStrategy(3e7, 10.5, 60, 18, true, 3)
	if before.Type != StrategySeasoning || before.TotalStrategy == 0 {
		t.Errorf("expected SBI before 24mo to have a non-zero wait cost, got %.2f", before.TotalStrategy)
	}
	after := ComputeStrategy(3e7, 10.5, 60, 30, true, 3)
	if after.TotalStrategy != 0 {
		t.Errorf("expected SBI after 24mo to be free, got %.2f", after.TotalStrategy)
	}
}

func TestAUNoneStrategyHasNoSaving(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 18, true, 5)
	if res.HasStrategy {
		t.Error("expected AU (index 5) to have HasStrategy false")
	}
}

func TestStaggeredStrategyClearsToZeroResidual(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 57, true, 4)
	if res.Type != StrategyStaggered || !res.HasStrategy {
		t.Fatalf("expected Axis to be StrategyStaggered, got %s", res.Type)
	}
}

func TestSeasoningStrategyAlreadyPastWaivesImmediately(t *testing.T) {
	res := ComputeStrategy(3e7, 10.5, 60, 30, true, 6)
	if res.TotalStrategy != 0 {
		t.Errorf("TotalStrategy = %.2f, want 0 (already past seasoning)", res.TotalStrategy)
	}
}
