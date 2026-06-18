package game

import (
	"escape-room/models"
	"log"
	"sync"
	"time"
)

type GameRunner struct {
	scriptID  uint
	elapsed   int
	duration  int
	triggerAt int
	triggered bool
	ticker    *time.Ticker
	stopChan  chan struct{}
	mu        sync.Mutex
	paused    bool
}

type GameManager struct {
	runners map[uint]*GameRunner
	mu      sync.RWMutex
}

var (
	manager *GameManager
	once    sync.Once
)

func GetManager() *GameManager {
	once.Do(func() {
		manager = &GameManager{
			runners: make(map[uint]*GameRunner),
		}
	})
	return manager
}

func (gm *GameManager) StartGame(scriptID uint) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.runners[scriptID]; exists {
		return nil
	}

	var script models.Script
	if err := models.GetDB().First(&script, scriptID).Error; err != nil {
		return err
	}

	runner := &GameRunner{
		scriptID:  scriptID,
		elapsed:   0,
		duration:  script.Duration * 60,
		triggerAt: script.PressureTriggerAt,
		triggered: script.PressureTriggered,
		stopChan:  make(chan struct{}),
	}

	gm.runners[scriptID] = runner
	go runner.run()

	log.Printf("[GameManager] Script %d started, duration: %ds, pressure trigger at %ds remaining",
		scriptID, runner.duration, runner.triggerAt)

	return nil
}

func (gm *GameManager) PauseGame(scriptID uint) {
	gm.mu.RLock()
	runner, exists := gm.runners[scriptID]
	gm.mu.RUnlock()

	if !exists {
		return
	}

	runner.mu.Lock()
	runner.paused = true
	runner.mu.Unlock()

	log.Printf("[GameManager] Script %d paused at %ds", scriptID, runner.elapsed)
}

func (gm *GameManager) ResumeGame(scriptID uint) {
	gm.mu.RLock()
	runner, exists := gm.runners[scriptID]
	gm.mu.RUnlock()

	if !exists {
		return
	}

	runner.mu.Lock()
	runner.paused = false
	runner.mu.Unlock()

	log.Printf("[GameManager] Script %d resumed at %ds", scriptID, runner.elapsed)
}

func (gm *GameManager) StopGame(scriptID uint) {
	gm.mu.Lock()
	runner, exists := gm.runners[scriptID]
	if exists {
		close(runner.stopChan)
		delete(gm.runners, scriptID)
	}
	gm.mu.Unlock()

	log.Printf("[GameManager] Script %d stopped", scriptID)
}

func (gm *GameManager) ResetGame(scriptID uint) {
	gm.StopGame(scriptID)

	models.GetDB().Model(&models.Script{}).Where("id = ?", scriptID).
		Update("pressure_triggered", false)

	log.Printf("[GameManager] Script %d reset", scriptID)
}

func (gm *GameManager) GetGameStatus(scriptID uint) map[string]interface{} {
	gm.mu.RLock()
	runner, exists := gm.runners[scriptID]
	gm.mu.RUnlock()

	if !exists {
		var script models.Script
		if err := models.GetDB().First(&script, scriptID).Error; err != nil {
			return nil
		}
		return map[string]interface{}{
			"running":             false,
			"paused":              false,
			"elapsed_seconds":     0,
			"remaining_seconds":   script.Duration * 60,
			"total_seconds":       script.Duration * 60,
			"pressure_triggered":  script.PressureTriggered,
			"pressure_trigger_at": script.PressureTriggerAt,
		}
	}

	runner.mu.Lock()
	defer runner.mu.Unlock()

	remaining := runner.duration - runner.elapsed
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"running":             true,
		"paused":              runner.paused,
		"elapsed_seconds":     runner.elapsed,
		"remaining_seconds":   remaining,
		"total_seconds":       runner.duration,
		"pressure_triggered":  runner.triggered,
		"pressure_trigger_at": runner.triggerAt,
	}
}

func (gm *GameManager) TriggerPressure(scriptID uint) error {
	var script models.Script
	if err := models.GetDB().Preload("Rooms.Devices").Preload("Rooms.Audios").
		First(&script, scriptID).Error; err != nil {
		return err
	}

	tx := models.GetDB().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, room := range script.Rooms {
		for _, device := range room.Devices {
			if device.IsPressure {
				if err := tx.Model(&device).Update("status", models.DeviceOn).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
		for _, audio := range room.Audios {
			if audio.IsPressure {
				if err := tx.Model(&audio).Updates(map[string]interface{}{
					"status": models.AudioPlaying,
					"volume": 80,
				}).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	if err := tx.Model(&script).Update("pressure_triggered", true).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	gm.mu.RLock()
	runner, exists := gm.runners[scriptID]
	gm.mu.RUnlock()
	if exists {
		runner.mu.Lock()
		runner.triggered = true
		runner.mu.Unlock()
	}

	log.Printf("[GameManager] Pressure triggered for script %d", scriptID)
	return nil
}

func (gm *GameManager) ResetPressureStatus(scriptID uint) {
	gm.mu.RLock()
	runner, exists := gm.runners[scriptID]
	gm.mu.RUnlock()

	if exists {
		runner.mu.Lock()
		runner.triggered = false
		runner.mu.Unlock()
	}
}

func (gr *GameRunner) run() {
	gr.ticker = time.NewTicker(1 * time.Second)
	defer gr.ticker.Stop()

	for {
		select {
		case <-gr.stopChan:
			return
		case <-gr.ticker.C:
			gr.mu.Lock()
			if gr.paused {
				gr.mu.Unlock()
				continue
			}

			gr.elapsed++
			remaining := gr.duration - gr.elapsed

			if !gr.triggered && remaining <= gr.triggerAt && remaining > 0 {
				gr.triggered = true
				gr.mu.Unlock()
				go func() {
					if err := GetManager().TriggerPressure(gr.scriptID); err != nil {
						log.Printf("[GameManager] Trigger pressure failed for script %d: %v", gr.scriptID, err)
					}
				}()
			} else {
				gr.mu.Unlock()
			}

			if remaining <= 0 {
				gr.mu.Lock()
				gr.paused = true
				gr.mu.Unlock()
				log.Printf("[GameManager] Script %d time's up", gr.scriptID)
			}
		}
	}
}
