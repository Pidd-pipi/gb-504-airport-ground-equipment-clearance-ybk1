package service

import (
	"fmt"
	"strconv"
	"time"

	"groundclearance/internal/constants"
	"groundclearance/internal/model"
	"groundclearance/internal/repository"

	"gorm.io/gorm"
)

// UnitValidity describes the re-inspection freshness of one ground unit for a
// given flight risk level.
type UnitValidity struct {
	UnitID       uint64     `json:"unit_id"`
	UnitCode     string     `json:"unit_code"`
	UnitState    string     `json:"unit_state"`
	RiskLevel    string     `json:"risk_level"`
	ValidityFrom *time.Time `json:"validity_from"`
	WindowHours  int        `json:"window_hours"`
	Expired      bool       `json:"expired"`
	Reason       string     `json:"reason"`
}

func riskWindow(riskLevel string) time.Duration {
	if riskLevel == constants.RiskHigh || riskLevel == constants.RiskCritical {
		return constants.ReinspectionWindowHighRisk
	}
	return constants.ReinspectionWindowStandard
}

// validityAnchor prefers the latest passing re-inspection; a failed outcome
// must not refresh validity. When no re-inspection exists the pre-shift
// inspection time is used.
func validityAnchor(unit *model.GroundUnit, passed map[uint64]*model.ReinspectionRecord) *time.Time {
	if record, ok := passed[unit.ID]; ok {
		return &record.InspectedAt
	}
	return unit.LastInspectionAt
}

// EvaluateUnitValidity checks one unit against the risk-specific window. Units
// not in available state are governed by the state rules instead of freshness.
func EvaluateUnitValidity(unit *model.GroundUnit, passed map[uint64]*model.ReinspectionRecord, riskLevel string, now time.Time) UnitValidity {
	window := riskWindow(riskLevel)
	result := UnitValidity{UnitID: unit.ID, UnitCode: unit.UnitCode, UnitState: unit.State,
		RiskLevel: riskLevel, WindowHours: int(window.Hours())}
	if unit.State != constants.UnitAvailable {
		return result
	}
	anchor := validityAnchor(unit, passed)
	result.ValidityFrom = anchor
	if anchor == nil {
		result.Expired = true
		result.Reason = fmt.Sprintf("设备 %s 没有复检有效期内的检查记录", unit.UnitCode)
		return result
	}
	if now.Sub(*anchor) > window {
		result.Expired = true
		result.Reason = fmt.Sprintf("设备 %s 复检已超 %d 小时有效期（航班风险等级%s）", unit.UnitCode, result.WindowHours, riskLevel)
	}
	return result
}

// LoadAssignedUnitValidity evaluates every unit assigned to a turnaround
// inside the given transaction. Invalid id references are reported so callers
// can block clearance decisions.
func LoadAssignedUnitValidity(tx *gorm.DB, unitRepo *repository.GroundUnitRepository,
	reinspectionRepo *repository.ReinspectionRepository, turnaround *model.Turnaround) ([]UnitValidity, []uint64, error) {
	unitIDs := make([]uint64, 0, len(turnaround.GroundUnitIDs))
	for _, rawID := range turnaround.GroundUnitIDs {
		id, err := strconv.ParseUint(rawID, 10, 64)
		if err != nil || id == 0 {
			return nil, nil, fmt.Errorf("turnaround contains an invalid ground unit id: %s", rawID)
		}
		unitIDs = append(unitIDs, id)
	}
	passed, err := reinspectionRepo.FindLatestPassedByGroundUnitIDsTx(tx, unitIDs)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	validities := make([]UnitValidity, 0, len(unitIDs))
	expired := make([]uint64, 0)
	for _, id := range unitIDs {
		unit, findErr := unitRepo.FindByIDTx(tx, id)
		if findErr != nil {
			return nil, nil, findErr
		}
		validity := EvaluateUnitValidity(unit, passed, turnaround.RiskLevel, now)
		if validity.Expired {
			expired = append(expired, id)
		}
		validities = append(validities, validity)
	}
	return validities, expired, nil
}
