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

func (s *InternalNotificationService) GetUnreadNotificationsCount(userID uint) (count int64, err error) {
	return count, s.DB.Model(&InternalNotification{}).Where("user_id = ? AND read = ?", userID, false).Count(&count).Error
}

// MarkAsRead 将通知标记为已读
func (s *InternalNotificationService) MarkAsRead(notificationID uint) error {
	return s.DB.Model(&InternalNotification{}).Where("id = ?", notificationID).Update("read", true).Error
}

// MarkAsRead 将通知标记为已读
func (s *InternalNotificationService) MarkAllAsRead(userID uint) error {
	return s.DB.Model(&InternalNotification{}).Where("user_id = ?", userID).Update("read", true).Error
}

// GetPaginatedNotifications 获取用户的分页通知
func (s *InternalNotificationService) GetPaginatedNotifications(
	userID uint,
	page, size int,
	filter string,
	titleKeyword, contentKeyword string,
	startTime, endTime time.Time,
) ([]InternalNotification, int64, error) {
	var notifications []InternalNotification
	var total int64

	db := s.DB.Model(&InternalNotification{}).Where("user_id = ?", userID)
	// 是否已读过滤
	if filter == "read" {
		db = db.Where("read = ?", true)
	} else if filter == "unread" {
		db = db.Where("read = ?", false)
	}
	// 标题模糊匹配
	if titleKeyword != "" {
		db = db.Where("title LIKE ?", "%"+titleKeyword+"%")
	}
	// 内容模糊匹配
	if contentKeyword != "" {
		db = db.Where("content LIKE ?", "%"+contentKeyword+"%")
	}
	// 时间范围过滤
	if !startTime.IsZero() && !endTime.IsZero() {
		db = db.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	} else if !startTime.IsZero() {
		db = db.Where("created_at >= ?", startTime)
	} else if !endTime.IsZero() {
		db = db.Where("created_at <= ?", endTime)
	}
	// 查询总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// 分页查询
	err := db.Offset((page - 1) * size).Limit(size).Order("created_at DESC").Find(&notifications).Error
	return notifications, total, err
}

func (s *InternalNotificationService) GetOne(userID uint, notificationID uint) (InternalNotification, error) {
	var notification InternalNotification
	return notification, s.DB.Where("user_id = ? AND id = ?", userID, notificationID).First(&notification).Error
}

func (s *InternalNotificationService) Delete(userID uint, notificationID uint) error {
	return s.DB.Where("user_id = ? AND id = ?", userID, notificationID).Delete(&InternalNotification{}).Error
}
