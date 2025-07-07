package handlers

import (
	"VoiceSculptor/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mock user for auth injection
var mockUser = &models.User{
	ID:    1,
	Email: "test@example.com",
}

// 模拟中间件注入用户
func mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", mockUser)
		c.Next()
	}
}

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.Assistant{})
	assert.NoError(t, err)

	// 插入测试用户
	err = db.Create(mockUser).Error
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mockAuthMiddleware())

	h := &Handlers{db: db}

	api := router.Group("/api")
	{
		api.POST("/assistant", h.CreateAssistant)
		api.GET("/assistants", h.ListAssistants)
		api.GET("/assistant/:id", h.GetAssistant)
		api.PUT("/assistant/:id", h.UpdateAssistant)
		api.DELETE("/assistant/:id", h.DeleteAssistant)
	}

	return router, db
}

func TestAssistantCRUD(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create assistant
	createPayload := map[string]interface{}{
		"name":         "编程导师",
		"systemPrompt": "你是一位耐心的Go语言讲师",
		"instruction":  "请简洁回答，不要加废话",
		"personaTag":   "go_teacher",
		"temperature":  0.7,
		"maxTokens":    100,
	}
	body, _ := json.Marshal(createPayload)

	req := httptest.NewRequest("POST", "/api/assistant", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var created models.Assistant
	err := json.Unmarshal(resp.Body.Bytes(), &created)
	assert.NoError(t, err)
	assert.Equal(t, "编程导师", created.Name)

	// Get list
	req = httptest.NewRequest("GET", "/api/assistants", nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var list []models.Assistant
	err = json.Unmarshal(resp.Body.Bytes(), &list)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// Update assistant
	update := map[string]interface{}{
		"name": "高级编程导师",
	}
	updateBody, _ := json.Marshal(update)
	req = httptest.NewRequest("PUT", "/api/assistant/"+toStr(created.ID), bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Get single assistant
	req = httptest.NewRequest("GET", "/api/assistant/"+toStr(created.ID), nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	var updated models.Assistant
	err = json.Unmarshal(resp.Body.Bytes(), &updated)
	assert.NoError(t, err)
	assert.Equal(t, "高级编程导师", updated.Name)

	// Delete assistant
	req = httptest.NewRequest("DELETE", "/api/assistant/"+toStr(created.ID), nil)
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

// 辅助字符串转化
func toStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
