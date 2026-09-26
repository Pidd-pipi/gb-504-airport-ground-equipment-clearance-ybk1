package service

import (
	"testing"
	"time"

	"groundclearance/internal/constants"
	"groundclearance/internal/model"
)

func TestInspectionWindowForRisk(t *testing.T) {
	if got := constants.InspectionWindowForRisk(constants.RiskHigh); got != 8*time.Hour {
		t.Fatalf("high risk window: got %v want 8h", got)
	}
	if got := constants.InspectionWindowForRisk(constants.RiskCritical); got != 8*time.Hour {
		t.Fatalf("critical risk window: got %v want 8h", got)
	}
	for _, risk := range []string{constants.RiskLow, constants.RiskMedium, ""} {
		if got := constants.InspectionWindowForRisk(risk); got != 24*time.Hour {
			t.Fatalf("risk %q window: got %v want 24h", risk, got)
		}
	}
}

func TestMarkUnitInspection(t *testing.T) {
	now := time.Now()
	unit := &model.GroundUnit{UnitCode: "TUG-001"}
	markUnitInspection(unit, 24*time.Hour, now)
	if !unit.InspectionExpired || unit.InspectionDueAt != nil {
		t.Fatal("missing inspection record must be treated as expired")
	}
	fresh := now.Add(-2 * time.Hour)
	unit.LastInspectionAt = &fresh
	markUnitInspection(unit, 24*time.Hour, now)
	if unit.InspectionExpired {
		t.Fatal("fresh inspection must stay valid")
	}
	if unit.InspectionDueAt == nil || !unit.InspectionDueAt.Equal(fresh.Add(24*time.Hour)) {
		t.Fatalf("unexpected due time: %v", unit.InspectionDueAt)
	}
	stale := now.Add(-9 * time.Hour)
	unit.LastInspectionAt = &stale
	markUnitInspection(unit, 24*time.Hour, now)
	if unit.InspectionExpired {
		t.Fatal("9h old inspection must stay valid under the 24h window")
	}
	markUnitInspection(unit, 8*time.Hour, now)
	if !unit.InspectionExpired {
		t.Fatal("9h old inspection must expire under the 8h high-risk window")
	}
}

func TestOverdueUnits(t *testing.T) {
	now := time.Now()
	fresh := now.Add(-3 * time.Hour)
	stale := now.Add(-30 * time.Hour)
	units := []model.GroundUnit{
		{UnitCode: "TUG-001", LastInspectionAt: &fresh},
		{UnitCode: "GPU-002", LastInspectionAt: &stale},
		{UnitCode: "BLT-003"},
	}
	overdue := overdueUnits(units, constants.RiskMedium, now)
	if len(overdue) != 2 || overdue[0].UnitCode != "GPU-002" || overdue[1].UnitCode != "BLT-003" {
		t.Fatalf("unexpected overdue units: %#v", overdueUnitCodes(overdue))
	}
	if got := overdueUnits(units[:1], constants.RiskHigh, now); len(got) != 0 {
		t.Fatalf("fresh unit must not be overdue: %#v", overdueUnitCodes(got))
	}
}

func TestHighRiskActiveUnits(t *testing.T) {
	rows := []model.Turnaround{
		{ID: 1, RiskLevel: constants.RiskHigh, GroundUnitIDs: model.JSONList{"7", "8"}},
		{ID: 2, RiskLevel: constants.RiskLow, GroundUnitIDs: model.JSONList{"9"}},
		{ID: 3, RiskLevel: constants.RiskCritical, GroundUnitIDs: model.JSONList{"9"}},
	}
	highRisk := highRiskActiveUnits(rows)
	if !highRisk[7] || !highRisk[8] || !highRisk[9] {
		t.Fatalf("expected units 7, 8 and 9 to be high risk: %#v", highRisk)
	}
	if highRisk[10] {
		t.Fatal("unassigned unit must not be high risk")
	}
	if got := unitInspectionWindow(7, highRisk); got != 8*time.Hour {
		t.Fatalf("high risk unit window: got %v want 8h", got)
	}
	if got := unitInspectionWindow(10, highRisk); got != 24*time.Hour {
		t.Fatalf("default unit window: got %v want 24h", got)
	}
}
