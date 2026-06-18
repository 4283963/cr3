package handlers

import (
	"escape-room/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AudioHandler struct{}

func NewAudioHandler() *AudioHandler {
	return &AudioHandler{}
}

func (h *AudioHandler) GetAll(c *gin.Context) {
	roomID := c.Query("room_id")
	audioType := c.Query("type")
	var audios []models.Audio
	db := models.GetDB()
	if roomID != "" {
		db = db.Where("room_id = ?", roomID)
	}
	if audioType != "" {
		db = db.Where("type = ?", audioType)
	}
	if err := db.Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": audios})
}

func (h *AudioHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": audio})
}

func (h *AudioHandler) Create(c *gin.Context) {
	var audio models.Audio
	if err := c.ShouldBindJSON(&audio); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audio.Status = models.AudioStopped
	if audio.Volume == 0 {
		audio.Volume = 50
	}
	if err := models.GetDB().Create(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": audio})
}

func (h *AudioHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}

	if err := c.ShouldBindJSON(&audio); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audio.ID = uint(id)
	if err := models.GetDB().Save(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": audio})
}

func (h *AudioHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	if err := models.GetDB().Delete(&models.Audio{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "audio deleted"})
}

type PlayAudioRequest struct {
	Volume int  `json:"volume"`
	Loop   bool `json:"loop"`
}

func (h *AudioHandler) Play(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var req PlayAudioRequest
	_ = c.ShouldBindJSON(&req)

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}

	audio.Status = models.AudioPlaying
	if req.Volume > 0 {
		audio.Volume = req.Volume
	}
	if req.Loop {
		audio.Loop = req.Loop
	}

	if err := models.GetDB().Save(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "audio playing", "data": audio})
}

func (h *AudioHandler) Pause(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}

	audio.Status = models.AudioPaused
	if err := models.GetDB().Save(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "audio paused", "data": audio})
}

func (h *AudioHandler) Stop(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}

	audio.Status = models.AudioStopped
	if err := models.GetDB().Save(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "audio stopped", "data": audio})
}

type VolumeRequest struct {
	Volume int `json:"volume" binding:"required,min=0,max=100"`
}

func (h *AudioHandler) SetVolume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio id"})
		return
	}

	var req VolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var audio models.Audio
	if err := models.GetDB().First(&audio, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audio not found"})
		return
	}

	audio.Volume = req.Volume
	if err := models.GetDB().Save(&audio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "volume updated", "data": audio})
}

func (h *AudioHandler) SetRoomBGMVolume(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req VolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := models.GetDB().Model(&models.Audio{}).
		Where("room_id = ? AND type = ?", roomID, models.AudioBGM).
		Update("volume", req.Volume).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var audios []models.Audio
	if err := models.GetDB().Where("room_id = ? AND type = ?", roomID, models.AudioBGM).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bgm volume updated", "data": audios})
}

func (h *AudioHandler) ControlRoomBGM(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	action := c.Query("action")
	var status models.AudioStatus
	switch action {
	case "play":
		status = models.AudioPlaying
	case "pause":
		status = models.AudioPaused
	case "stop":
		status = models.AudioStopped
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action, must be play/pause/stop"})
		return
	}

	if err := models.GetDB().Model(&models.Audio{}).
		Where("room_id = ? AND type = ?", roomID, models.AudioBGM).
		Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var audios []models.Audio
	if err := models.GetDB().Where("room_id = ? AND type = ?", roomID, models.AudioBGM).Find(&audios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bgm " + action + "d", "data": audios})
}
