//go:build oracle

package util

import (
	"gorm.io/driver/oracle"
	"gorm.io/gorm"
)

func createDatabaseInstance(cfg *gorm.Config, driver, dsn string) (*gorm.DB, error) {
	return gorm.Open(oracle.Open(dsn), cfg)
}
