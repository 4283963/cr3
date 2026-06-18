package handlers

import (
	"escape-room/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ScriptHandler struct{}

func NewScriptHandler() *ScriptHandler {
	return &ScriptHandler{}
}

func (h *ScriptHandler) GetAll(c *gin.Context) {
	var scripts []models.Script
	if err := models.GetDB().Preload("Rooms").Find(&scripts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": scripts})
}

func (h *ScriptHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().Preload("Rooms.Devices").Preload("Rooms.Audios").First(&script, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": script})
}

func (h *ScriptHandler) Create(c *gin.Context) {
	var script models.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	script.Status = models.ScriptIdle
	if err := models.GetDB().Create(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": script})
}

func (h *ScriptHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().First(&script, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	script.ID = uint(id)
	if err := models.GetDB().Save(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": script})
}

func (h *ScriptHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	if err := models.GetDB().Delete(&models.Script{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "script deleted"})
}

func (h *ScriptHandler) Start(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().First(&script, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	now := time.Now()
	script.Status = models.ScriptRunning
	script.StartTime = &now

	tx := models.GetDB().Begin()
	if err := tx.Save(&script).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var rooms []models.Room
	if err := tx.Where("script_id = ?", id).Find(&rooms).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, room := range rooms {
		if err := tx.Model(&models.Device{}).Where("room_id = ?", room.ID).Update("status", models.DeviceOff).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Model(&models.Audio{}).Where("room_id = ?", room.ID).Update("status", models.AudioStopped).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "script started", "data": script})
}

func (h *ScriptHandler) Pause(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().First(&script, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	script.Status = models.ScriptPaused
	if err := models.GetDB().Save(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "script paused", "data": script})
}

func (h *ScriptHandler) Stop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	var script models.Script
	if err := models.GetDB().First(&script, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	script.Status = models.ScriptCompleted
	script.StartTime = nil
	if err := models.GetDB().Save(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "script stopped", "data": script})
}

func (h *ScriptHandler) Reset(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid script id"})
		return
	}

	tx := models.GetDB().Begin()

	var script models.Script
	if err := tx.First(&script, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "script not found"})
		return
	}

	script.Status = models.ScriptIdle
	script.StartTime = nil
	if err := tx.Save(&script).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var rooms []models.Room
	if err := tx.Where("script_id = ?", id).Find(&rooms).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, room := range rooms {
		if err := tx.Model(&models.Device{}).Where("room_id = ?", room.ID).Update("status", models.DeviceOff).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Model(&models.Audio{}).Where("room_id = ?", room.ID).Updates(map[string]interface{}{
			"status": models.AudioStopped,
			"volume": 50,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "script reset"})
}
