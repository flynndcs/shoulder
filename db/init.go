package db

import "gorm.io/gorm"

func InitDb(db *gorm.DB) {
	db.AutoMigrate(&Accretion{})
}
