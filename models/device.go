package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceStatus string

const (
	DeviceOff   DeviceStatus = "off"
	DeviceOn    DeviceStatus = "on"
	DeviceError DeviceStatus = "error"
)

type DeviceType string

const (
	DeviceSwitch   DeviceType = "switch"
	DeviceLight    DeviceType = "light"
	DeviceLock     DeviceType = "lock"
	DeviceSensor   DeviceType = "sensor"
	DeviceMotor    DeviceType = "motor"
	DeviceCustom   DeviceType = "custom"
)

type Device struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RoomID      uint           `gorm:"not null;index" json:"room_id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Type        DeviceType     `gorm:"not null;size:20" json:"type"`
	Status      DeviceStatus   `gorm:"default:off;size:20" json:"status"`
	GPIO        string         `gorm:"size:50" json:"gpio"`
	Description string         `gorm:"size:500" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
