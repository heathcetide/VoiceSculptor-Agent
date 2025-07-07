package handlers

import (
	"VoiceSculptor/internal/models"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	_ = db.AutoMigrate(&models.Assistant{}, &models.User{}, &models.UserCredential{})
	return db
}

func TestChatHandler_StartAndStopVoiceCall(t *testing.T) {
	db := setupTestDB()

	// 初始化测试数据
	user := &models.User{DisplayName: "test-user"}
	db.Create(user)

	cred := &models.UserCredential{
		UserID:      user.ID,
		Name:        "test-cred",
		LLMProvider: "openai",
		LLMApiKey:   "83fb4b9faddc98a274664b5bd4141aa7.6nNum7w223OgxqV3",
		Quota:       10000,
	}
	db.Create(cred)

	assistant := &models.Assistant{
		UserID:       int64(user.ID),
		Name:         "优雅导师",
		SystemPrompt: "你是一位温柔的导师。",
		Instruction:  "请耐心解答用户的问题。",
		PersonaTag:   "mentor",
		Temperature:  0.6,
		MaxTokens:    150,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Create(assistant)

	// 初始化 Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	h := NewHandlers(db)
	router.Use(func(c *gin.Context) {
		c.Set("user", user)
		c.Set("credential", cred)
	})
	router.POST("/api/chat/start", h.Chat)
	router.POST("/api/chat/stop", h.StopChat)

	// 启动通话
	startPayload, _ := json.Marshal(map[string]interface{}{
		"assistantId": assistant.ID,
	})
	req := httptest.NewRequest("POST", "/api/chat/start", bytes.NewBuffer(startPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	t.Log("Start response:", w.Body.String())

	// 等待模拟会话
	time.Sleep(3 * time.Minute)

	// 停止通话
	stopReq := httptest.NewRequest("POST", "/api/chat/stop", nil)
	stopW := httptest.NewRecorder()
	router.ServeHTTP(stopW, stopReq)

	assert.Equal(t, http.StatusOK, stopW.Code)
	t.Log("Stop response:", stopW.Body.String())

	var stopResp map[string]string
	_ = json.Unmarshal(stopW.Body.Bytes(), &stopResp)
	assert.Equal(t, "通话已关闭", stopResp["message"])
}
