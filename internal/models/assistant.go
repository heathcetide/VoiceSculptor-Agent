package models

import "time"

// Assistant 表示一个自定义的 AI 助手
type Assistant struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64     `json:"userId" gorm:"index"`
	Name         string    `json:"name" gorm:"index"`
	SystemPrompt string    `json:"systemPrompt"`
	Instruction  string    `json:"instruction"`
	PersonaTag   string    `json:"personaTag"`
	Temperature  float32   `json:"temperature"`
	MaxTokens    int       `json:"maxTokens"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ChatLog 表示一次调用记录（可用于扣费、审计等）
type ChatLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64     `json:"userId"`
	AssistantID  int64     `json:"assistantId"`
	Input        string    `json:"input"`
	Output       string    `json:"output"`
	PromptTokens int       `json:"promptTokens"`
	OutputTokens int       `json:"outputTokens"`
	CreatedAt    time.Time `json:"createdAt"`
}
