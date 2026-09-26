package service

import (
	"strconv"
	"time"

	"groundclearance/internal/constants"
	"groundclearance/internal/model"
)

// markUnitInspection fills the transient re-inspection status on a ground unit.
// A missing inspection record is treated as expired: the unit cannot prove a
// valid pre-shift check.
func markUnitInspection(unit *model.GroundUnit, window time.Duration, now time.Time) {
	unit.InspectionWindowHours = int(window / time.Hour)
	unit.InspectionDueAt = nil
	unit.InspectionExpired = true
	if unit.LastInspectionAt == nil {
		return
	}
	dueAt := unit.LastInspectionAt.Add(window)
	unit.InspectionDueAt = &dueAt
	unit.InspectionExpired = now.After(dueAt)
}

// highRiskActiveUnits indexes units assigned to active high or critical risk
// turnarounds, which tightens their re-inspection window to 8 hours.
func highRiskActiveUnits(turnarounds []model.Turnaround) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, row := range turnarounds {
		if row.RiskLevel != constants.RiskHigh && row.RiskLevel != constants.RiskCritical {
			continue
		}
		for _, rawID := range row.GroundUnitIDs {
			if id, err := strconv.ParseUint(rawID, 10, 64); err == nil && id != 0 {
				result[id] = true
			}
		}
	}
	return result
}

// unitInspectionWindow resolves the validity window for one unit on the
// equipment desk: 8 hours while the unit feeds a high or critical risk
// turnaround, otherwise the default 24 hours.
func unitInspectionWindow(unitID uint64, highRisk map[uint64]bool) time.Duration {
	if highRisk[unitID] {
		return constants.InspectionWindowForRisk(constants.RiskCritical)
	}
	return constants.InspectionWindowForRisk(constants.RiskLow)
}

// overdueUnits returns the assigned units whose re-inspection validity has
// expired under the turnaround risk window.
func overdueUnits(units []model.GroundUnit, riskLevel string, now time.Time) []model.GroundUnit {
	window := constants.InspectionWindowForRisk(riskLevel)
	overdue := make([]model.GroundUnit, 0)
	for _, unit := range units {
		probe := unit
		markUnitInspection(&probe, window, now)
		if probe.InspectionExpired {
			overdue = append(overdue, probe)
		}
	}
	return overdue
}

// overdueUnitCodes renders unit codes for error messages and audit detail.
func overdueUnitCodes(units []model.GroundUnit) []string {
	codes := make([]string, 0, len(units))
	for _, unit := range units {
		codes = append(codes, unit.UnitCode)
	}
	return codes
}
