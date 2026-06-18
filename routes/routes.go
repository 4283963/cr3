package routes

import (
	"escape-room/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	scriptHandler := handlers.NewScriptHandler()
	roomHandler := handlers.NewRoomHandler()
	deviceHandler := handlers.NewDeviceHandler()
	audioHandler := handlers.NewAudioHandler()
	controlHandler := handlers.NewControlHandler()

	api := r.Group("/api/v1")
	{
		scripts := api.Group("/scripts")
		{
			scripts.GET("", scriptHandler.GetAll)
			scripts.GET("/:id", scriptHandler.GetByID)
			scripts.POST("", scriptHandler.Create)
			scripts.PUT("/:id", scriptHandler.Update)
			scripts.DELETE("/:id", scriptHandler.Delete)
			scripts.POST("/:id/start", scriptHandler.Start)
			scripts.POST("/:id/pause", scriptHandler.Pause)
			scripts.POST("/:id/stop", scriptHandler.Stop)
			scripts.POST("/:id/reset", scriptHandler.Reset)
		}

		rooms := api.Group("/rooms")
		{
			rooms.GET("", roomHandler.GetAll)
			rooms.GET("/:id", roomHandler.GetByID)
			rooms.POST("", roomHandler.Create)
			rooms.PUT("/:id", roomHandler.Update)
			rooms.DELETE("/:id", roomHandler.Delete)
		}

		devices := api.Group("/devices")
		{
			devices.GET("", deviceHandler.GetAll)
			devices.GET("/:id", deviceHandler.GetByID)
			devices.POST("", deviceHandler.Create)
			devices.PUT("/:id", deviceHandler.Update)
			devices.DELETE("/:id", deviceHandler.Delete)
			devices.POST("/:id/trigger", deviceHandler.Trigger)
			devices.POST("/:id/toggle", deviceHandler.Toggle)
		}

		audios := api.Group("/audios")
		{
			audios.GET("", audioHandler.GetAll)
			audios.GET("/:id", audioHandler.GetByID)
			audios.POST("", audioHandler.Create)
			audios.PUT("/:id", audioHandler.Update)
			audios.DELETE("/:id", audioHandler.Delete)
			audios.POST("/:id/play", audioHandler.Play)
			audios.POST("/:id/pause", audioHandler.Pause)
			audios.POST("/:id/stop", audioHandler.Stop)
			audios.POST("/:id/volume", audioHandler.SetVolume)
		}

		control := api.Group("/control")
		{
			control.GET("/scripts/:scriptId/status", controlHandler.GetScriptStatus)
			control.GET("/rooms/:roomId/status", controlHandler.GetRoomStatus)
			control.POST("/rooms/:roomId/devices/trigger", controlHandler.TriggerRoomDevice)
			control.POST("/rooms/:roomId/devices/trigger-all", controlHandler.TriggerAllRoomDevices)
			control.POST("/rooms/:roomId/reset", controlHandler.ResetRoom)
			control.POST("/rooms/:roomId/bgm/volume", audioHandler.SetRoomBGMVolume)
			control.POST("/rooms/:roomId/bgm/control", audioHandler.ControlRoomBGM)
		}
	}

	return r
}
