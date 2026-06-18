package handlers

import (
	"escape-room/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ControlHandler struct{}

func NewControlHandler() *ControlHandler {
	return &ControlHandler{}
}

type TriggerRoomDeviceRequest struct {
	DeviceName string              `json:"device_name"`
	DeviceID   uint                `json:"device_id"`
	Status     models.DeviceStatus `json:"status" binding:"required"`
}

func (h *ControlHandler) TriggerRoomDevice(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req TriggerRoomDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != models.DeviceOn && req.Status != models.DeviceOff && req.Status != models.DeviceError {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device status"})
		return
	}

	var device models.Device
	db := models.GetDB().Where("room_id = ?", roomID)
	if req.DeviceID > 0 {
		db = db.Where("id = ?", req.DeviceID)
	} else if req.DeviceName != "" {
		db = db.Where("name = ?", req.DeviceName)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id or device_name is required"})
		return
	}

	if err := db.First(&device).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found in this room"})
		return
	}

	device.Status = req.Status
	if err := models.GetDB().Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "device triggered successfully",
		"data":    device,
	})
}

type TriggerAllRoomDevicesRequest struct {
	Status     models.DeviceStatus `json:"status" binding:"required"`
	DeviceType models.DeviceType   `json:"device_type"`
}

func (h *ControlHandler) TriggerAllRoomDevices(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req TriggerAllRoomDevicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status != models.DeviceOn && req.Status != models.DeviceOff && req.Status != models.DeviceError {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device status"})
		return
	}

	db := models.GetDB().Model(&models.Device{}).Where("room_id = ?", roomID)
	if req.DeviceType != "" {
		db = db.Where("type = ?", req.DeviceType)
	}

	if err := db.Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var devices []models.Device
	query := models.GetDB().Where("room_id = ?", roomID)
	if req.DeviceType != "" {
		query = query.Where("type = ?", req.DeviceType)
	}
	if err := query.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all devices triggered",
		"data":    devices,
	})
}

func (h *ControlHandler) GetRoomStatus(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var room models.Room
	if err := models.GetDB().Preload("Devices").Preload("Audios").First(&room, roomID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

func (h *ControlHandler) GetScriptStatus(c *gin.Context) {
	scriptID, err := strconv.ParseUint(c.Param("scriptId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().Preload("Rooms.Devices").Preload("Rooms.Audios").First(&script, scriptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": script})
}

type ResetRoomRequest struct {
	ResetDevices bool `json:"reset_devices" default:"true"`
	ResetAudios  bool `json:"reset_audios" default:"true"`
}

func (h *ControlHandler) ResetRoom(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req ResetRoomRequest
	req.ResetDevices = true
	req.ResetAudios = true
	_ = c.ShouldBindJSON(&req)

	tx := models.GetDB().Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
		return
	}

	var checkRoom models.Room
	if err := tx.First(&checkRoom, roomID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	if req.ResetDevices {
		if err := tx.Model(&models.Device{}).Where("room_id = ?", roomID).Update("status", models.DeviceOff).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.ResetAudios {
		if err := tx.Model(&models.Audio{}).Where("room_id = ?", roomID).Updates(map[string]interface{}{
			"status": models.AudioStopped,
			"volume": 50,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var room models.Room
	if err := models.GetDB().Preload("Devices").Preload("Audios").First(&room, roomID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "room reset successfully",
		"data":    room,
	})
}
