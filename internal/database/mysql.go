package database

import (
	"backend/internal/config"
	"backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenMySQL(cfg config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.RefreshToken{})
}
