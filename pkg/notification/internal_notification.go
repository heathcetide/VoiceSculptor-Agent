package notification

import (
	"gorm.io/gorm"
	"time"
)

// InternalNotification 站内通知
type InternalNotification struct {
	ID        uint      `json:"id" gorm:"primaryKey"` // 通知 ID
	UserID    uint      `json:"user_id"`              // 用户 ID
	Title     string    `json:"title"`                // 通知标题
	Content   string    `json:"content"`              // 通知内容
	Read      bool      `json:"read"`                 // 是否已读
	CreatedAt time.Time `json:"created_at"`           // 创建时间
}

// InternalNotificationService 站内通知服务
type InternalNotificationService struct {
	DB *gorm.DB // 数据库实例
}

// NewInternalNotificationService 创建站内通知服务实例
func NewInternalNotificationService(db *gorm.DB) *InternalNotificationService {
	return &InternalNotificationService{DB: db}
}

// Send 发送站内通知
func (s *InternalNotificationService) Send(userID uint, title, content string) error {
	notification := InternalNotification{
		UserID:    userID,
		Title:     title,
		Content:   content,
		Read:      false,
		CreatedAt: time.Now(),
	}

	// 将通知存储到数据库
	return s.DB.Create(&notification).Error
}

// GetUnreadNotifications 获取用户的未读通知
func (s *InternalNotificationService) GetUnreadNotifications(userID uint) ([]InternalNotification, error) {
	var notifications []InternalNotification
	err := s.DB.Where("user_id = ? AND read = ?", userID, false).Find(&notifications).Error
	return notifications, err
}

// MarkAsRead 将通知标记为已读
func (s *InternalNotificationService) MarkAsRead(notificationID uint) error {
	return s.DB.Model(&InternalNotification{}).Where("id = ?", notificationID).Update("read", true).Error
}
