package models

import (
	"time"

	"gorm.io/gorm"
)

type ScriptStatus string

const (
	ScriptIdle      ScriptStatus = "idle"
	ScriptRunning   ScriptStatus = "running"
	ScriptPaused    ScriptStatus = "paused"
	ScriptCompleted ScriptStatus = "completed"
)

type Script struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"not null;size:100" json:"name"`
	Description        string         `gorm:"size:500" json:"description"`
	Status             ScriptStatus   `gorm:"default:idle;size:20" json:"status"`
	Duration           int            `gorm:"default:60" json:"duration"`
	PressureTriggerAt  int            `gorm:"default:600" json:"pressure_trigger_at"`
	PressureTriggered  bool           `gorm:"default:false" json:"pressure_triggered"`
	Rooms              []Room         `gorm:"foreignKey:ScriptID" json:"rooms,omitempty"`
	StartTime          *time.Time     `json:"start_time,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}
