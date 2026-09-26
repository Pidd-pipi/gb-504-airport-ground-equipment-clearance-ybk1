package model

import "time"

// GroundUnit is an airside vehicle or powered support unit.
type GroundUnit struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UnitCode         string     `gorm:"size:40;uniqueIndex;not null" json:"unit_code"`
	Name             string     `gorm:"size:120;not null" json:"name"`
	UnitType         string     `gorm:"size:40;not null;index" json:"unit_type"`
	Stand            string     `gorm:"size:30;not null;index" json:"stand"`
	State            string     `gorm:"size:30;not null;default:available;index;check:chk_ground_units_state,state IN ('available','inspection','blocked','retired')" json:"state"`
	LastInspectionAt *time.Time `json:"last_inspection_at"`
	Notes            string     `gorm:"type:text" json:"notes"`
	Version          int        `gorm:"not null;default:1" json:"version"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Computed view fields populated by the service layer; not persisted.
	LatestReinspection *ReinspectionRecord `gorm:"-" json:"latest_reinspection,omitempty"`
	// Validity flags are evaluated for the 24h standard window and the 8h
	// high/critical risk window so the equipment page can mark expiry.
	Expired24h bool `gorm:"-" json:"expired_24h"`
	Expired8h  bool `gorm:"-" json:"expired_8h"`
}

func (GroundUnit) TableName() string { return "ground_units" }
