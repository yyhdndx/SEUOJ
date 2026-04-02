package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"seu-oj-backend/internal/config"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
}
