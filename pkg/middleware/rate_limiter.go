package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginmiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"sync"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	Rate string `json:"rate"` // 限流规则，如 "10-S" 表示每秒 10 个请求
}

var (
	rateLimiterConfig = &RateLimiterConfig{Rate: "10-S"} // 默认限流规则
	rateLimiterMutex  sync.RWMutex                       // 用于保护限流配置的读写
)

// SetRateLimiterConfig 动态更新限流配置
func SetRateLimiterConfig(config RateLimiterConfig) {
	rateLimiterMutex.Lock()
	defer rateLimiterMutex.Unlock()
	rateLimiterConfig = &config
}

// GetRateLimiterConfig 获取当前限流配置
func GetRateLimiterConfig() RateLimiterConfig {
	rateLimiterMutex.RLock()
	defer rateLimiterMutex.RUnlock()
	return *rateLimiterConfig
}

// RateLimiterMiddleware 动态限流中间件
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前限流配置
		config := GetRateLimiterConfig()

		// 创建限流器
		rateLimit, err := limiter.NewRateFromFormatted(config.Rate)
		if err != nil {
			panic(err)
		}

		// 使用内存存储
		store := memory.NewStore()

		// 创建限流中间件
		middleware := ginmiddleware.NewMiddleware(limiter.New(store, rateLimit))

		// 执行限流中间件
		middleware(c)
	}
}
