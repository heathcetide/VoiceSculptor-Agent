package handlers

import (
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/client"
	"VoiceSculptor/pkg/config"
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/response"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var sessions sync.Map

var sessionMessages sync.Map // sessionId -> *[]string

type ChatRequest struct {
	AssistantID  int64   `json:"assistantId" binding:"required"`
	SystemPrompt string  `json:"systemPrompt"`
	Instruction  string  `json:"instruction"`
	Speaker      string  `json:"speaker"`
	Language     string  `json:"language"`
	ApiKey       string  `json:"apiKey"`
	ApiSecret    string  `json:"apiSecret"`
	Speed        float32 `json:"speed"`
	Volume       int     `json:"volume"`
	PersonaTag   string  `json:"personaTag"`
	Temperature  float32 `json:"temperature"`
	MaxTokens    int     `json:"maxTokens"`
}

// ChatResponse 只是响应状态（实际处理是语音流）
type ChatResponse struct {
	Message string `json:"message"`
}

type ChatSessionMap struct {
	AssistantID int64
	SdkClient   *client.Client
}

func (h *Handlers) Chat(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	cred, err := models.GetUserCredentialByApiSecretAndApiKey(h.db, req.ApiKey, req.ApiSecret)
	if err != nil {
		response.Fail(c, "Database error: "+err.Error(), nil)
		return
	}
	if cred == nil {
		response.Fail(c, "Secret or key is invalid or wrong. Please check again.", nil)
		return
	}

	// 构建 Prompt 配置
	prompt := client.PromptConfig{
		SystemPrompt:   req.SystemPrompt,
		Instruction:    req.Instruction,
		PersonaTag:     req.PersonaTag,
		Temperature:    req.Temperature,
		MaxTokens:      req.MaxTokens,
		HistoryEnabled: false,
	}

	// 初始化语音 SDK 客户端
	cfg := client.Config{
		ICEURL:     config.GlobalConfig.RustPbxUrl + "/iceservers",
		ServerAddr: config.GlobalConfig.RustPbxWebSocketURL + "/call/webrtc",
		ASR: client.AsrConfig{
			Provider:  cred.AsrProvider,
			AppId:     cred.AsrAppID,
			SecretId:  cred.AsrSecretID,
			SecretKey: cred.AsrSecretKey,
			Language:  req.Language,
		},
		TTS: client.TtsConfig{
			Provider:  cred.TtsProvider,
			Speaker:   req.Speaker,
			AppId:     cred.TTSAppID,
			SecretId:  cred.TTSSecretID,
			SecretKey: cred.TTSSecretKey,
			Speed:     req.Speed,
			Volume:    req.Volume,
		},
		OpenAIKey: cred.LLMApiKey,
		Prompt:    prompt,
	}
	ctx := context.Background()
	sessionID := uuid.New().String()
	sdkClient, err := client.NewClient(ctx, cfg)
	// Chat 启动时
	sessions.Store(sessionID, ChatSessionMap{
		req.AssistantID,
		sdkClient,
	})
	if err != nil {
		response.Fail(c, "SDK 初始化失败", gin.H{"error": "SDK 初始化失败: " + err.Error()})
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

	response.Success(c, "语音通话已启动", gin.H{
		"message":   "语音通话已启动",
		"sessionId": sessionID,
	})
}

func (h *Handlers) StopChat(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		response.Fail(c, "not find param sessionId ", nil)
		return
	}
	// StopChat 时
	if val, ok := sessions.Load(sessionID); ok {
		sdk := val.(ChatSessionMap)
		sdk.SdkClient.Close()
		sessions.Delete(sessionID)
		// 2. 保存历史记录
		if val, ok := sessionMessages.Load(sessionID); ok {
			msgs := *val.(*[]string)
			sessionMessages.Delete(sessionID)

			content := strings.Join(msgs, "\n") // 整段拼接
			user := models.CurrentUser(c)
			if user == nil {
				response.Fail(c, "User is not logged in.", nil)
				return
			}
			// 写入数据库
			log := models.ChatSessionLog{
				SessionID:   sessionID,
				UserID:      user.ID,
				AssistantID: sdk.AssistantID,
				Content:     content,
				CreatedAt:   time.Now(),
			}
			h.db.Create(&log)
		}

		response.Success(c, "success", ChatResponse{
			Message: "语音通话已停止",
		})
		return
	}
	response.Fail(c, "未找到活动通话", nil)
}

func (h *Handlers) ChatStream(c *gin.Context) {
	sessionId := c.Query("sessionId")
	if sessionId == "" {
		response.Fail(c, "not find param sessionId ", nil)
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// 获取当前用户对应的 sdkClient
	value, ok := sessions.Load(sessionId)
	if !ok {
		c.SSEvent("error", "No active chat session")
		return
	}
	sdk := value.(ChatSessionMap)

	ch := sdk.SdkClient.SSEChannel()

	// 推送消息
	for {
		select {
		case <-c.Request.Context().Done():
			return // 客户端关闭连接
		case msg := <-ch:
			// 存储消息
			if val, ok := sessionMessages.Load(sessionId); ok {
				msgs := val.(*[]string)
				*msgs = append(*msgs, msg)
			} else {
				sessionMessages.Store(sessionId, &[]string{msg})
			}
			c.SSEvent("message", msg)
			c.Writer.Flush()
		}
	}
}

func (h *Handlers) getChatSessionLog(c *gin.Context) {
	// 获取当前登录用户
	user := models.CurrentUser(c)
	if user == nil {
		// 用户未登录
		response.Fail(c, "User is not logged in.", nil)
		return
	}

	// 获取分页参数，默认为每页10条记录
	pageSize := c.DefaultQuery("pageSize", "10")
	cursor := c.DefaultQuery("cursor", "") // 游标ID（即最后一条记录的ID）

	pageSizeInt, _ := strconv.Atoi(pageSize)

	// 如果 pageSize <= 0，返回错误
	if pageSizeInt <= 0 {
		response.Fail(c, "Invalid pageSize", nil)
		return
	}

	// 创建查询条件
	var logs []models.ChatSessionLog
	var query *gorm.DB

	// 如果有游标，按游标进行查询
	if cursor != "" {
		// 解析游标为 ID（int64）
		cursorID, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			response.Fail(c, "Invalid cursor", nil)
			return
		}

		// 查询大于游标的记录
		query = h.db.Where("user_id = ? AND id > ?", user.ID, cursorID).Order("id asc").Limit(pageSizeInt)
	} else {
		// 第一次查询，直接按时间倒序查询
		query = h.db.Where("user_id = ?", user.ID).Order("id asc").Limit(pageSizeInt)
	}

	// 执行查询
	if err := query.Find(&logs).Error; err != nil {
		// 查询失败
		response.Fail(c, "Failed to fetch chat logs", nil)
		return
	}

	// 如果查询不到记录，返回空数据
	if len(logs) == 0 {
		response.Success(c, "No more chat logs", nil)
		return
	}
	// 创建一个切片来存放查询结果
	var result []map[string]interface{}
	for _, log := range logs {
		// 截取 content 中的第一个换行符之前的部分
		content := log.Content
		if index := strings.Index(content, "\n"); index != -1 {
			content = content[:index] // 截取内容
		}

		// 将每条记录格式化并添加到 result 切片中
		result = append(result, map[string]interface{}{
			"id":        log.ID,
			"content":   content,
			"createdAt": log.CreatedAt,
		})
	}

	// 获取下一页的游标
	var nextCursor int64
	if len(logs) > 0 {
		nextCursor = logs[len(logs)-1].ID
	}

	// 返回成功的响应，并带上下一页游标
	response.Success(c, "Fetched chat session logs successfully", map[string]interface{}{
		"logs":        result,
		"nextCursor":  nextCursor,
		"hasMoreData": len(logs) == pageSizeInt, // 判断是否还有更多数据
	})
}

func (h *Handlers) getChatSessionLogDetail(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Fail(c, "not find param id ", nil)
		return
	}
	var log models.ChatSessionLog
	if err := h.db.Where("id = ?", id).First(&log).Error; err != nil {
		response.Fail(c, "Failed to fetch chat log", nil)
		return
	}
	response.Success(c, "Fetched chat log successfully", log)
}
