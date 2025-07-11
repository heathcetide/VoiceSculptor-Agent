package handlers

import (
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/notification"
	"VoiceSculptor/pkg/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// GetUnReadNotificationCount get user unread notification count
func (h *Handlers) handleUnReadNotificationCount(c *gin.Context) {
	user := models.CurrentUser(c)

	users, err := models.GetUserByEmail(h.db, user.Email)
	if err != nil {
		response.AbortWithStatus(c, http.StatusUnauthorized)
		return
	}
	unreadNotificationCount, err := notification.NewInternalNotificationService(h.db).GetUnreadNotificationsCount(users.ID)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusInternalServerError, err)
		return
	}
	response.Success(c, "success", unreadNotificationCount)
}

// ListNotifications list user notifications
func (h *Handlers) handleListNotifications(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
	}
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	var (
		pageInt  int
		sizeInt  int
		filterBy = c.Query("filter")  // read / unread
		title    = c.Query("title")   // 按标题查询
		content  = c.Query("content") // 按内容查询
		layout   = "2006-01-02T15:04:05Z07:00"
		startStr = c.Query("start_time") // 开始时间
		endStr   = c.Query("end_time")   // 结束时间
		start    time.Time
		end      time.Time
	)

	_, _ = fmt.Sscanf(page, "%d", &pageInt)
	_, _ = fmt.Sscanf(size, "%d", &sizeInt)

	if startStr != "" {
		start, _ = time.Parse(layout, startStr)
	}
	if endStr != "" {
		end, _ = time.Parse(layout, endStr)
	}

	service := notification.NewInternalNotificationService(h.db)
	notifications, total, err := service.GetPaginatedNotifications(
		user.ID,
		pageInt,
		sizeInt,
		filterBy,
		title,
		content,
		start,
		end,
	)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusInternalServerError, err)
		return
	}
	response.Success(c, "success", gin.H{
		"list":  notifications,
		"total": total,
		"page":  pageInt,
		"size":  sizeInt,
	})
}

// AllNotifications mark all notifications as read
func (h *Handlers) handleAllNotifications(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
	}
	err := notification.NewInternalNotificationService(h.db).MarkAllAsRead(user.ID)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusInternalServerError, err)
		return
	}
	response.Success(c, "already mark all notifications", nil)
}

// handleMarkNotificationAsRead 将指定通知标记为已读
func (h *Handlers) handleMarkNotificationAsRead(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
		return
	}

	// 获取路径参数中的 notification ID
	idStr := c.Param("id")
	var notificationID uint
	_, err := fmt.Sscanf(idStr, "%d", &notificationID)
	if err != nil {
		response.AbortWithStatus(c, http.StatusBadRequest)
		return
	}

	_, err = notification.NewInternalNotificationService(h.db).GetOne(user.ID, notificationID)
	if err != nil {
		response.Fail(c, "You don't have permission to flag this message.", nil)
		return
	}

	// 调用服务层标记为已读
	err = notification.NewInternalNotificationService(h.db).MarkAsRead(notificationID)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, "Notification marked as read", nil)
}

func (h *Handlers) handleDeleteNotification(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
		return
	}
	var notificationID uint
	_, err := fmt.Sscanf(c.Param("id"), "%d", &notificationID)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusBadRequest, err)
		return
	}
	err = notification.NewInternalNotificationService(h.db).Delete(user.ID, notificationID)
	if err != nil {
		response.AbortWithStatusJSON(c, http.StatusInternalServerError, err)
		return
	}
	response.Success(c, "Notification deleted", nil)
}
