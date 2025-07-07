//go:build pg

package util

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func createDatabaseInstance(cfg *gorm.Config, driver, dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), cfg)
}
