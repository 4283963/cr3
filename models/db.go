package models

import (
	"escape-room/config"
	"log"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func InitDB(cfg *config.Config) *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		err = db.AutoMigrate(&Script{}, &Room{}, &Device{}, &Audio{})
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		log.Println("Database initialized successfully")
	})
	return db
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Panic("database has not been initialized, call InitDB first")
	}
	return db
}
