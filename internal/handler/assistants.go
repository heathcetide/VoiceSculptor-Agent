package handlers

import (
	"VoiceSculptor/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

// 创建新的助手
func (h *Handlers) CreateAssistant(c *gin.Context) {
	var input struct {
		Name         string  `json:"name" binding:"required"`
		SystemPrompt string  `json:"systemPrompt"`
		Instruction  string  `json:"instruction"`
		PersonaTag   string  `json:"personaTag"`
		Temperature  float32 `json:"temperature"`
		MaxTokens    int     `json:"maxTokens"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	user, ok := c.MustGet("user").(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}

	assistant := models.Assistant{
		UserID:       int64(user.ID),
		Name:         input.Name,
		SystemPrompt: input.SystemPrompt,
		Instruction:  input.Instruction,
		PersonaTag:   input.PersonaTag,
		Temperature:  input.Temperature,
		MaxTokens:    input.MaxTokens,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&assistant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建助手失败"})
		return
	}
	c.JSON(http.StatusOK, assistant)
}

// 查询当前用户所有助手
func (h *Handlers) ListAssistants(c *gin.Context) {
	user, ok := c.MustGet("user").(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var list []models.Assistant
	if err := h.db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, list)
}

// 查询单个助手
func (h *Handlers) GetAssistant(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var assistant models.Assistant
	if err := h.db.First(&assistant, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "助手不存在"})
		return
	}
	c.JSON(http.StatusOK, assistant)
}

// 更新助手
func (h *Handlers) UpdateAssistant(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input models.Assistant
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效请求"})
		return
	}
	if err := h.db.Model(&models.Assistant{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":          input.Name,
			"system_prompt": input.SystemPrompt,
			"instruction":   input.Instruction,
			"persona_tag":   input.PersonaTag,
			"temperature":   input.Temperature,
			"max_tokens":    input.MaxTokens,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// 删除助手
func (h *Handlers) DeleteAssistant(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.db.Delete(&models.Assistant{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
