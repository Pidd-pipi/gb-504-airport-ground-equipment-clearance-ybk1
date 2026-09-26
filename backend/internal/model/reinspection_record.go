package model

import "time"

// ReinspectionRecord is a shift re-check of one ground unit. It refreshes the
// unit's re-inspection validity anchor without changing the unit state.
type ReinspectionRecord struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	GroundUnitID uint64    `gorm:"not null;index" json:"ground_unit_id"`
	Result       string    `gorm:"size:20;not null;index;check:chk_reinspections_result,result IN ('passed','failed')" json:"result"`
	Evidence     JSONList  `gorm:"type:jsonb" json:"evidence"`
	Remark       string    `gorm:"type:text" json:"remark"`
	InspectorID  uint64    `gorm:"not null;index" json:"inspector_id"`
	InspectedAt  time.Time `gorm:"not null;index" json:"inspected_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ReinspectionRecord) TableName() string { return "reinspection_records" }
