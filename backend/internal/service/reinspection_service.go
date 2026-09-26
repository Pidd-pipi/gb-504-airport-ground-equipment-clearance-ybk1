package service

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
	"groundclearance/internal/constants"
	"groundclearance/internal/model"
	"groundclearance/internal/repository"
	"groundclearance/internal/util"
)

type ReinspectionService struct {
	db       *gorm.DB
	repo     *repository.ReinspectionRepository
	unitRepo *repository.GroundUnitRepository
	logger   *slog.Logger
}

func NewReinspectionService(db *gorm.DB, repo *repository.ReinspectionRepository,
	unitRepo *repository.GroundUnitRepository, logger *slog.Logger) *ReinspectionService {
	return &ReinspectionService{db: db, repo: repo, unitRepo: unitRepo, logger: logger}
}

// Register records a shift re-inspection. The equipment state is deliberately
// left untouched; only the validity anchor and audit trail change.
func (s *ReinspectionService) Register(groundUnitID, operatorID uint64, result string,
	evidence []string, remark string, actor AuditContext) (*model.ReinspectionRecord, error) {
	if !constants.In(constants.ReinspectionResultValues, result) {
		return nil, util.NewAppError(constants.CodeValidationFailed, "invalid reinspection result")
	}
	evidence, err := normalizeEvidence(evidence)
	if err != nil {
		return nil, err
	}
	remark = strings.TrimSpace(remark)
	var record *model.ReinspectionRecord
	err = s.db.Transaction(func(tx *gorm.DB) error {
		unit, findErr := s.unitRepo.FindByIDTx(tx, groundUnitID)
		if findErr != nil {
			if errors.Is(findErr, repository.ErrNotFound) {
				return util.NewAppError(constants.CodeNotFound, constants.MsgNotFound)
			}
			return findErr
		}
		if unit.State == constants.UnitRetired {
			return util.NewAppError(constants.CodeStateConflict, "retired equipment cannot be reinspected")
		}
		now := time.Now()
		record = &model.ReinspectionRecord{
			GroundUnitID: unit.ID, Result: result, Evidence: model.JSONList(evidence),
			Remark: remark, InspectorID: operatorID, InspectedAt: now,
		}
		if err := s.repo.CreateTx(tx, record); err != nil {
			return err
		}
		return persistTransitionAudit(tx, actor, "GROUND_UNIT_REINSPECTION", "ground-units", unit.ID, map[string]any{
			"unit_code": unit.UnitCode, "unit_state_unchanged": unit.State,
			"result": result, "evidence": evidence, "remark": remark,
			"inspected_at": now, "inspector_id": operatorID,
			"reinspection_id": record.ID,
		})
	})
	if err != nil {
		return nil, err
	}
	s.logger.Info(constants.LogGroundUnitReinspected, "ground_unit_id", groundUnitID, "result", result, "operator_id", operatorID)
	return record, nil
}

func (s *ReinspectionService) ListByGroundUnit(groundUnitID uint64) ([]model.ReinspectionRecord, error) {
	if _, err := s.unitRepo.FindByID(groundUnitID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, util.NewAppError(constants.CodeNotFound, constants.MsgNotFound)
		}
		return nil, err
	}
	return s.repo.ListByGroundUnit(s.db, groundUnitID)
}
