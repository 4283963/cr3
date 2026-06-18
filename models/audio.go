package models

import (
	"time"

	"gorm.io/gorm"
)

type AudioStatus string

const (
	AudioStopped AudioStatus = "stopped"
	AudioPlaying AudioStatus = "playing"
	AudioPaused  AudioStatus = "paused"
)

type AudioType string

const (
	AudioBGM    AudioType = "bgm"
	AudioEffect AudioType = "effect"
	AudioVoice  AudioType = "voice"
)

type Audio struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	RoomID     uint           `gorm:"not null;index" json:"room_id"`
	Name       string         `gorm:"not null;size:100" json:"name"`
	Type       AudioType      `gorm:"not null;size:20" json:"type"`
	FilePath   string         `gorm:"size:500" json:"file_path"`
	Status     AudioStatus    `gorm:"default:stopped;size:20" json:"status"`
	Volume     int            `gorm:"default:50" json:"volume"`
	Loop       bool           `gorm:"default:false" json:"loop"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
