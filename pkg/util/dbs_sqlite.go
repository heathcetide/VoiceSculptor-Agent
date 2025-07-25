//go:build !mysql && !pg

package util

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func createDatabaseInstance(cfg *gorm.Config, driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), cfg)
	case "pg":
		return gorm.Open(postgres.Open(dsn), cfg)
	}
	if dsn == "" {
		dsn = "file::memory:"
	}
	return gorm.Open(sqlite.Open(dsn), cfg)
}
