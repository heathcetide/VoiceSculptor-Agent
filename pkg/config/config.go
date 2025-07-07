package config

import (
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/notification"
	"VoiceSculptor/pkg/util"
	"log"
	"os"
)

// config/config.go
type Config struct {
	DBDriver         string `env:"DB_DRIVER"`
	DSN              string `env:"DSN"`
	Log              logger.LogConfig
	Mail             notification.MailConfig
	Addr             string `env:"ADDR"`
	Mode             string `env:"MODE"`
	DocsPrefix       string `env:"DOCS_PREFIX"`
	APIPrefix        string `env:"API_PREFIX"`
	AdminPrefix      string `env:"ADMIN_PREFIX"`
	AuthPrefix       string `env:"AUTH_PREFIX"`
	SessionSecret    string `env:"SESSION_SECRET"`
	SecretExpireDays string `env:"SESSION_EXPIRE_DAYS"`
}

var GlobalConfig *Config

func Load() error {
	// 1. 根据环境加载 .env 文件
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // 默认使用开发环境
	}
	err := util.LoadEnv(env)
	if err != nil {
		log.Printf("Failed to load .env file: %v", err)
	}

	// 2. 加载全局配置
	GlobalConfig = &Config{
		DBDriver:         util.GetEnv("DB_DRIVER"),
		DSN:              util.GetEnv("DSN"),
		Addr:             util.GetEnv("ADDR"),
		Mode:             util.GetEnv("MODE"),
		DocsPrefix:       util.GetEnv("DOCS_PREFIX"),
		APIPrefix:        util.GetEnv("API_PREFIX"),
		AdminPrefix:      util.GetEnv("ADMIN_PREFIX"),
		AuthPrefix:       util.GetEnv("AUTH_PREFIX"),
		SecretExpireDays: util.GetEnv("SESSION_EXPIRE_DAYS"),
		SessionSecret:    util.GetEnv("SESSION_SECRET"),
		Log: logger.LogConfig{
			Level:      util.GetEnv("LOG_LEVEL"),
			Filename:   util.GetEnv("LOG_FILENAME"),
			MaxSize:    int(util.GetIntEnv("LOG_MAX_SIZE")),
			MaxAge:     int(util.GetIntEnv("LOG_MAX_AGE")),
			MaxBackups: int(util.GetIntEnv("LOG_MAX_BACKUPS")),
		},
		Mail: notification.MailConfig{
			Host:     util.GetEnv("MAIL_HOST"),
			Username: util.GetEnv("MAIL_USERNAME"),
			Password: util.GetEnv("MAIL_PASSWORD"),
			Port:     util.GetIntEnv("MAIL_PORT"),
			From:     util.GetEnv("MAIL_FROM"),
		},
	}
	return nil
}
