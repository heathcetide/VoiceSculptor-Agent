package notification

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 初始化测试用的 SQLite 数据库
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&InternalNotification{})
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

func TestInternalNotificationService_Send(t *testing.T) {
	// 初始化测试数据库
	db := setupTestDB()

	// 创建站内通知服务实例
	service := NewInternalNotificationService(db)

	// 测试发送通知
	err := service.Send(1, "Test Title", "This is a test notification.")
	if err != nil {
		t.Errorf("Failed to send internal notification: %v", err)
	} else {
		t.Log("Internal notification sent successfully!")
	}

	// 验证通知是否存储到数据库
	var notification InternalNotification
	result := db.First(&notification, "user_id = ?", 1)
	if result.Error != nil {
		t.Errorf("Failed to find notification: %v", result.Error)
	} else {
		t.Logf("Notification found: %+v", notification)
	}
}

func TestInternalNotificationService_GetUnreadNotifications(t *testing.T) {
	// 初始化测试数据库
	db := setupTestDB()

	// 创建站内通知服务实例
	service := NewInternalNotificationService(db)

	// 添加测试数据
	notifications := []InternalNotification{
		{UserID: 1, Title: "Test 1", Content: "Content 1", Read: false, CreatedAt: time.Now()},
		{UserID: 1, Title: "Test 2", Content: "Content 2", Read: true, CreatedAt: time.Now()},
		{UserID: 2, Title: "Test 3", Content: "Content 3", Read: false, CreatedAt: time.Now()},
	}
	for _, n := range notifications {
		db.Create(&n)
	}

	// 测试获取未读通知
	unreadNotifications, err := service.GetUnreadNotifications(1)
	if err != nil {
		t.Errorf("Failed to get unread notifications: %v", err)
	} else {
		t.Logf("Unread notifications: %+v", unreadNotifications)
	}

	// 验证未读通知数量
	if len(unreadNotifications) != 1 {
		t.Errorf("Expected 1 unread notification, got %d", len(unreadNotifications))
	}
}

func TestInternalNotificationService_MarkAsRead(t *testing.T) {
	// 初始化测试数据库
	db := setupTestDB()

	// 创建站内通知服务实例
	service := NewInternalNotificationService(db)

	// 添加测试数据
	notification := InternalNotification{UserID: 1, Title: "Test", Content: "Content", Read: false, CreatedAt: time.Now()}
	db.Create(&notification)

	// 测试标记通知为已读
	err := service.MarkAsRead(notification.ID)
	if err != nil {
		t.Errorf("Failed to mark notification as read: %v", err)
	} else {
		t.Log("Notification marked as read!")
	}

	// 验证通知是否已读
	var updatedNotification InternalNotification
	db.First(&updatedNotification, "id = ?", notification.ID)
	if !updatedNotification.Read {
		t.Error("Notification was not marked as read")
	}
}
