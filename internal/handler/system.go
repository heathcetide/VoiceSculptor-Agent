package handlers

import (
	"VoiceSculptor/pkg/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UpdateRateLimiterConfig 更新限流配置
func (h *Handlers) UpdateRateLimiterConfig(c *gin.Context) {
	var config middleware.RateLimiterConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// 更新限流配置
	middleware.SetRateLimiterConfig(config)
	c.JSON(200, gin.H{"message": "rate limiter config updated"})
}

// HealthCheck 健康检查接口
func (h *Handlers) HealthCheck(c *gin.Context) {
	// 检查数据库连接
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database connection failed"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
		return
	}

	// 返回健康状态
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
