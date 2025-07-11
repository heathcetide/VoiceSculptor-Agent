package handlers

import (
	voiceSculptor "VoiceSculptor"
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/config"
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/response"
	"VoiceSculptor/pkg/util"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

// CreateAssistant create new assistant
func (h *Handlers) CreateAssistant(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	user := models.CurrentUser(c)

	assistant := models.Assistant{
		UserID:       user.ID,
		Name:         input.Name,
		Description:  input.Description,
		Icon:         input.Icon,
		SystemPrompt: "empty system prompt",
		Instruction:  "empty instruction",
		PersonaTag:   "mentor",
		Temperature:  0.6,
		MaxTokens:    150,
		JsSourceID:   strconv.FormatInt(util.SnowflakeUtil.NextID(), 20),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&assistant).Error; err != nil {
		response.Fail(c, fmt.Sprintf("create assistant %s failed", assistant.Name), nil)
		return
	}
	response.Success(c, fmt.Sprintf("create assistant %s successful", assistant.Name), assistant)
}

// ListAssistants 查询当前用户所有助手，并仅返回指定字段
func (h *Handlers) ListAssistants(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "unauthorized", "用户未登录")
		return
	}
	var list []struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
		JsSourceID  string `json:"jsSourceId"`
	}
	if err := h.db.Model(&models.Assistant{}).
		Where("user_id = ?", user.ID).
		Order("created_at desc").
		Select("id, name, description, icon, js_source_id").
		Find(&list).Error; err != nil {
		response.Fail(c, "select assistants failed", nil)
		return
	}

	response.Success(c, "select assistants successful", list)
}

// GetAssistant 查询单个助手
func (h *Handlers) GetAssistant(c *gin.Context) {
	user := models.CurrentUser(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var assistant models.Assistant
	if err := h.db.First(&assistant, id).Error; err != nil {
		response.Fail(c, "not found", "this assistant is not exist")
		return
	}
	if user.ID != assistant.UserID {
		response.Fail(c, "permission denied", "you are not allowed to access this assistant")
		return
	}
	response.Success(c, "select assistant successful", assistant)
}

// UpdateAssistant 更新助手
func (h *Handlers) UpdateAssistant(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "unauthorized", "用户未登录")
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var input struct {
		SystemPrompt string  `json:"systemPrompt"`
		Instruction  string  `json:"instruction"`
		PersonaTag   string  `json:"persona_tag"`
		Temperature  float32 `json:"temperature"`
		MaxTokens    int     `json:"maxTokens"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Fail(c, "invalid request", "parameter error")
		return
	}

	var assistant models.Assistant
	if err := h.db.First(&assistant, id).Error; err != nil {
		response.Fail(c, "not found", "Assistant does not exist.")
		return
	}

	if assistant.UserID != user.ID {
		response.Fail(c, "forbidden", "No permission to operate this assistant.")
		return
	}

	// 更新字段
	updateData := map[string]interface{}{
		"system_prompt": input.SystemPrompt,
		"instruction":   input.Instruction,
		"persona_tag":   input.PersonaTag,
		"temperature":   input.Temperature,
		"max_tokens":    input.MaxTokens,
		"updated_at":    time.Now(),
	}

	if err := h.db.Model(&assistant).Where("id = ?", id).Updates(updateData).Error; err != nil {
		response.Fail(c, "update failed", "更新失败")
		return
	}

	response.Success(c, "更新成功", assistant)
}

// DeleteAssistant 删除助手
func (h *Handlers) DeleteAssistant(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "unauthorized", "用户未登录")
		return
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var assistant models.Assistant
	if err := h.db.First(&assistant, id).Error; err != nil {
		response.Fail(c, "not found", "助手不存在")
		return
	}

	if assistant.UserID != user.ID {
		response.Fail(c, "forbidden", "无权限删除该助手")
		return
	}

	if err := h.db.Delete(&assistant, id).Error; err != nil {
		response.Fail(c, "delete failed", "删除失败")
		return
	}

	response.Success(c, "删除成功", nil)
}

func (h *Handlers) ServeVoiceSculptorLoaderJS(c *gin.Context) {
	jsSourceID := c.Param("id")
	var assistant models.Assistant
	err := h.db.Where("js_source_id = ?", jsSourceID).First(&assistant).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  http.StatusNotFound,
			"error": "assistant is not exists",
			"data":  nil,
		})
		return
	}

	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s%s", scheme, host, config.GlobalConfig.APIPrefix)

	tmpl, err := template.New("verification").Parse(voiceSculptor.AssistantJsModule)
	if err != nil {
		logger.Error("failed to parse verification template: ", zap.Error(err))
	}
	data := struct {
		BaseURL string
		Name    string
	}{
		BaseURL: baseURL,
		Name:    assistant.Name,
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		logger.Error("failed to render verification email: ", zap.Error(err))
	}

	c.Header("Content-Type", "application/javascript; charset=utf-8")
	c.String(http.StatusOK, body.String())
}
