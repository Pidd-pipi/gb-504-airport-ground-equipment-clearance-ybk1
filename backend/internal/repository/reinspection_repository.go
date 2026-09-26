package repository

import (
	"errors"
	"fmt"

	"groundclearance/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReinspectionRepository struct{ db *gorm.DB }

func NewReinspectionRepository(db *gorm.DB) *ReinspectionRepository {
	return &ReinspectionRepository{db: db}
}

func (r *ReinspectionRepository) CreateTx(tx *gorm.DB, record *model.ReinspectionRecord) error {
	if err := tx.Create(record).Error; err != nil {
		return fmt.Errorf("create reinspection record: %w", err)
	}
	return nil
}

func (r *ReinspectionRepository) FindLatestByGroundUnit(id uint64) (*model.ReinspectionRecord, error) {
	var record model.ReinspectionRecord
	err := r.db.Where("ground_unit_id = ?", id).Order("inspected_at DESC, id DESC").First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find latest reinspection: %w", err)
	}
	return &record, nil
}

func (r *ReinspectionRepository) FindLatestByGroundUnitIDsTx(tx *gorm.DB, ids []uint64) (map[uint64]*model.ReinspectionRecord, error) {
	result := make(map[uint64]*model.ReinspectionRecord, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var records []model.ReinspectionRecord
	if err := tx.Where("ground_unit_id IN ?", ids).Order("inspected_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list latest reinspections: %w", err)
	}
	for index := range records {
		record := records[index]
		if _, exists := result[record.GroundUnitID]; !exists {
			result[record.GroundUnitID] = &record
		}
	}
	return result, nil
}

// FindLatestPassedByGroundUnitIDsTx returns only passing re-inspections; failed
// outcomes must never extend the validity anchor.
func (r *ReinspectionRepository) FindLatestPassedByGroundUnitIDsTx(tx *gorm.DB, ids []uint64) (map[uint64]*model.ReinspectionRecord, error) {
	result := make(map[uint64]*model.ReinspectionRecord, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	var records []model.ReinspectionRecord
	if err := tx.Where("ground_unit_id IN ? AND result = ?", ids, "passed").
		Order("inspected_at DESC, id DESC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list latest passed reinspections: %w", err)
	}
	for index := range records {
		record := records[index]
		if _, exists := result[record.GroundUnitID]; !exists {
			result[record.GroundUnitID] = &record
		}
	}
	return result, nil
}

func (r *ReinspectionRepository) ListByGroundUnitTx(tx *gorm.DB, id uint64) ([]model.ReinspectionRecord, error) {
	var records []model.ReinspectionRecord
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("ground_unit_id = ?", id).Order("inspected_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list reinspections: %w", err)
	}
	return records, nil
}

func (r *ReinspectionRepository) ListByGroundUnit(db *gorm.DB, id uint64) ([]model.ReinspectionRecord, error) {
	var records []model.ReinspectionRecord
	if err := db.Where("ground_unit_id = ?", id).Order("inspected_at DESC, id DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list reinspections: %w", err)
	}
	return records, nil
}
