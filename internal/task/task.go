package task

import (
	"gorm.io/gorm"
	"time"
)

// StartOfflineChecker 启动用户离线检查任务
func StartOfflineChecker(db *gorm.DB) {
	ticker := time.NewTicker(2 * time.Minute) // 每 2 分钟检查一次
	defer ticker.Stop()

	for range ticker.C {
		checkOfflineUsers(db)
	}
}

// checkOfflineUsers check user status
func checkOfflineUsers(db *gorm.DB) {
	// 实现检查逻辑
	// 例如：查询最后活动时间超过一定阈值的用户，并标记为离线
}
