package models

import "time"

// Assistant 表示一个自定义的 AI 助手
type Assistant struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       uint      `json:"userId" gorm:"index"`
	Name         string    `json:"name" gorm:"index"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	SystemPrompt string    `json:"systemPrompt"`
	Instruction  string    `json:"instruction"`
	PersonaTag   string    `json:"personaTag"`
	Temperature  float32   `json:"temperature"`
	JsSourceID   string    `json:"jsSourceId"`
	MaxTokens    int       `json:"maxTokens"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ChatSessionLog 表示一次音频调用记录
type ChatSessionLog struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID   string    `json:"sessionId" gorm:"uniqueIndex"` // 与 SSE 保持一致
	UserID      uint      `json:"userId"`
	AssistantID int64     `json:"assistantId"`
	Content     string    `json:"content"` // 存储整段文本（[user]...\n[agent]...）
	CreatedAt   time.Time `json:"createdAt"`
}
