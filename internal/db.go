package internal

import (
	"log"

	database "shoulder/db"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDb(shoulderConfig ShoulderConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(shoulderConfig.PostgresConnString))
	if err != nil {
		log.Panicf("%s: %s", "Failed to connect to database", err)
	}

	database.InitDb(db)
	return db
}
