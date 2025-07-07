package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAIConfig 配置 OpenAI API 的相关信息
type OpenAIConfig struct {
	APIKey string
	URL    string
}

// OpenAIClient 是与 OpenAI API 通信的客户端
type OpenAIClient struct {
	Config OpenAIConfig
}

// OpenAIResponse 用于解析 OpenAI API 响应
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIRequest 是发送到 OpenAI API 的请求体
type OpenAIRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"` // "system", "user", "assistant"
		Content string `json:"content"`
	} `json:"messages"`
}

// NewOpenAIClient 创建一个新的 OpenAI 客户端实例
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		Config: OpenAIConfig{
			APIKey: apiKey,
			URL:    "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		},
	}
}

// GenerateText 调用 OpenAI API 生成文本
func (client *OpenAIClient) GenerateText(prompt string) (string, error) {
	// 构建请求体
	requestData := OpenAIRequest{
		Model: "glm-4", // 或者 "gpt-4" 具体的模型名称
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// 将请求体编码为 JSON
	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", client.Config.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.Config.APIKey)

	// 发起请求
	clientHTTP := &http.Client{Timeout: 30 * time.Second}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	var respData OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	// 返回生成的文本
	if len(respData.Choices) > 0 {
		return respData.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no choices found in the response")
}

// SendMessage 向用户发送消息并获得回应
func (client *OpenAIClient) SendMessage(content string) (string, error) {
	// 构建请求体
	requestData := OpenAIRequest{
		Model: "glm-4", // 或者 "gpt-4" 具体的模型名称
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "assistant", // 设置为 assistant 角色
				Content: content,
			},
		},
	}

	// 将请求体编码为 JSON
	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", client.Config.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+client.Config.APIKey)

	// 发起请求
	clientHTTP := &http.Client{Timeout: 30 * time.Second}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	var respData OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	// 返回生成的文本
	if len(respData.Choices) > 0 {
		return respData.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no choices found in the response")
}
