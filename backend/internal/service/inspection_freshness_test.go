package service

import (
	"testing"
	"time"

	"groundclearance/internal/constants"
	"groundclearance/internal/model"
)

func freshUnit(anchor time.Time) *model.GroundUnit {
	return &model.GroundUnit{ID: 1, UnitCode: "TUG-017", State: constants.UnitAvailable, LastInspectionAt: &anchor}
}

func TestEvaluateUnitValidityStandardWindow(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		anchor  time.Time
		risk    string
		expired bool
	}{
		{"medium within 24h", now.Add(-23 * time.Hour), constants.RiskMedium, false},
		{"medium beyond 24h", now.Add(-25 * time.Hour), constants.RiskMedium, true},
		{"high within 8h", now.Add(-7 * time.Hour), constants.RiskHigh, false},
		{"high beyond 8h but within 24h", now.Add(-10 * time.Hour), constants.RiskHigh, true},
		{"critical beyond 8h", now.Add(-9 * time.Hour), constants.RiskCritical, true},
		{"low beyond 8h still valid", now.Add(-10 * time.Hour), constants.RiskLow, false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			got := EvaluateUnitValidity(freshUnit(item.anchor), nil, item.risk, now)
			if got.Expired != item.expired {
				t.Fatalf("got expired=%v want %v (%s)", got.Expired, item.expired, got.Reason)
			}
		})
	}
}

func TestEvaluateUnitValidityAnchorRules(t *testing.T) {
	now := time.Now()

	// Missing anchor is always expired.
	missing := &model.GroundUnit{ID: 2, UnitCode: "GPU-204", State: constants.UnitAvailable}
	if !EvaluateUnitValidity(missing, nil, constants.RiskMedium, now).Expired {
		t.Fatal("unit without any inspection record must be expired")
	}

	// Non-available units are governed by state transitions, not freshness.
	old := now.Add(-72 * time.Hour)
	inspectionUnit := &model.GroundUnit{ID: 3, UnitCode: "BLT-088", State: constants.UnitInspection, LastInspectionAt: &old}
	if EvaluateUnitValidity(inspectionUnit, nil, constants.RiskCritical, now).Expired {
		t.Fatal("non-available units must not be flagged for re-inspection expiry")
	}

	// A passing re-inspection refreshes the anchor even if the pre-shift check is stale.
	stale := now.Add(-48 * time.Hour)
	passed := map[uint64]*model.ReinspectionRecord{
		1: {GroundUnitID: 1, Result: constants.ReinspectionPassed, InspectedAt: now.Add(-2 * time.Hour)},
	}
	if EvaluateUnitValidity(freshUnit(stale), passed, constants.RiskHigh, now).Expired {
		t.Fatal("passing re-inspection must refresh the validity anchor")
	}
}

func TestContainsReinspectionCondition(t *testing.T) {
	cases := map[string]bool{
		"TUG-017 完成复检通过后方可使用": true,
		"TUG-017 重新检查合格再投入":   true,
		"TUG-017 仅限低速牵引":      false,
		"":                    false,
	}
	for text, want := range cases {
		if got := containsReinspectionCondition(text); got != want {
			t.Fatalf("containsReinspectionCondition(%q)=%v want %v", text, got, want)
		}
	}
}
