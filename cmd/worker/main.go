package main

import (
	"VoiceSculptor/internal/task"
	"VoiceSculptor/pkg/config"
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/util"
	"flag"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	// 1. 解析命令行参数
	mode := flag.String("mode", "", "运行环境 (development, test, production)")
	flag.Parse()

	// 2. 设置环境变量
	if *mode != "" {
		os.Setenv("APP_ENV", *mode)
	}

	// 3. 加载全局配置
	if err := config.Load(); err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 4. 加载日志配置
	err := logger.Init(&config.GlobalConfig.Log, config.GlobalConfig.Mode)
	if err != nil {
		log.Fatalf("日志初始化失败: %v", err)
	}
	zap.L().Info("工作进程启动")

	// 5. 初始化数据库
	db, err := util.InitDatabase(os.Stdout, config.GlobalConfig.DBDriver, config.GlobalConfig.DSN)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 6. 启动后台任务
	go task.StartOfflineChecker(db) // 示例：启动用户离线检查任务

	// 7. 保持工作进程运行
	select {}
}
