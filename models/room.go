package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ScriptID    uint           `gorm:"not null;index" json:"script_id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Order       int            `gorm:"default:0" json:"order"`
	Devices     []Device       `gorm:"foreignKey:RoomID" json:"devices,omitempty"`
	Audios      []Audio        `gorm:"foreignKey:RoomID" json:"audios,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
