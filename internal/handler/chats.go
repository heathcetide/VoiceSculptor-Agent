package handlers

import (
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/client"
	"VoiceSculptor/pkg/logger"
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

var sessions sync.Map

type ChatRequest struct {
	AssistantID  int64  `json:"assistantId" binding:"required"`
	SystemPrompt string `json:"systemPrompt"`
	Instruction  string `json:"instruction"`
	Speaker      string `json:"speaker"`
	Language     string `json:"language"`
}

// ChatResponse 只是响应状态（实际处理是语音流）
type ChatResponse struct {
	Message string `json:"message"`
}

func (h *Handlers) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	user := &models.User{ID: 1, DisplayName: "test-user"}
	cred := &models.UserCredential{
		UserID:      user.ID,
		Name:        "test-cred",
		LLMProvider: "openai",
		LLMApiKey:   "83fb4b9faddc98a274664b5bd4141aa7.6nNum7w223OgxqV3",
		Quota:       10000,
	}
	//c.Set("user", user)
	//c.Set("credential", cred)
	//
	//cred := c.MustGet("credential").(*models.UserCredential)
	//user := c.MustGet("user").(*models.User)
	// 获取助手配置
	//var assistant models.Assistant
	//if err := h.db.First(&assistant, req.AssistantID).Error; err != nil {
	//	c.JSON(http.StatusNotFound, gin.H{"error": "助手不存在"})
	//	return
	//}

	// 构建 Prompt 配置
	prompt := client.PromptConfig{
		//SystemPrompt:   assistant.SystemPrompt,
		//Instruction:    assistant.Instruction,
		//PersonaTag:     assistant.PersonaTag,
		//Temperature:    assistant.Temperature,
		//MaxTokens:      assistant.MaxTokens,
		SystemPrompt:   req.SystemPrompt,
		Instruction:    req.Instruction,
		PersonaTag:     "mentor",
		Temperature:    0.6,
		MaxTokens:      150,
		HistoryEnabled: false, // 你可以改成 assistant 字段或用户设置
	}

	// 初始化语音 SDK 客户端
	cfg := client.Config{
		ICEURL:     "http://localhost:8080/iceservers",
		ServerAddr: "ws://localhost:8080/call/webrtc",
		ASR: client.AsrConfig{
			Provider:  "tencent",
			AppId:     "1325039295",
			SecretId:  "AKIDb4KNEWpvvx23yqdFh8Xlq9SeptmWadju",
			SecretKey: "Khx9wfaTNYiP5fFl7XsDYmqxwhLrfP1U",
			Language:  req.Language,
		},
		TTS: client.TtsConfig{
			Provider:  "tencent",
			Speaker:   req.Speaker, // 301030  "101016"
			AppId:     "1325039295",
			SecretId:  "AKIDb4KNEWpvvx23yqdFh8Xlq9SeptmWadju",
			SecretKey: "Khx9wfaTNYiP5fFl7XsDYmqxwhLrfP1U",
			Speed:     1.0,
			Volume:    5,
		},
		OpenAIKey: cred.LLMApiKey,
		Prompt:    prompt,
	}
	ctx := context.Background()
	sdkClient, err := client.NewClient(ctx, cfg)
	// Chat 启动时
	sessions.Store(user.ID, sdkClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SDK 初始化失败: " + err.Error()})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Error("StartCall panic", zap.Any("panic", r))
		}
	}()

	err = sdkClient.StartCall()
	if err != nil {
		logger.Error("StartCall error", zap.Error(err))
	}

	c.JSON(http.StatusOK, ChatResponse{
		Message: "语音通话已启动",
	})
}

func (h *Handlers) StopChat(c *gin.Context) {
	user := &models.User{ID: 1, DisplayName: "test-user"}
	// StopChat 时
	if val, ok := sessions.Load(user.ID); ok {
		sdk := val.(*client.Client)
		sdk.Close()
		sessions.Delete(user.ID)
		c.JSON(http.StatusOK, ChatResponse{
			Message: "语音通话已停止",
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "未找到活动通话"})
}
