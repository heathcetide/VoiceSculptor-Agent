package prompt

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestNewPromptManager(t *testing.T) {
	manager := newPromptManager()
	manager.registerPrompt(&Prompt{
		Name:        "story_generator",
		Description: "根据主角和场景生成故事开头",
		Arguments: []PromptArgument{
			{Name: "hero", Description: "故事主角", Required: true},
			{Name: "scene", Description: "故事场景", Required: true},
		},
	}, func(ctx context.Context, req *GetPromptRequest) (*GetPromptResult, error) {
		messages := []PromptMessage{
			{
				Role: "user",
				Content: TextContent{
					Type: "text",
					Text: fmt.Sprintf("主角是：%s，场景是：%s",
						req.Params.Arguments["hero"],
						req.Params.Arguments["scene"],
					),
				},
			},
		}
		return &GetPromptResult{
			Description: "生成故事开头的提示",
			Messages:    messages,
		}, nil
	})

	prompt, exists := manager.getPrompt("story_generator")
	if !exists {
		t.Errorf("prompt 'story_generator' should have been registered but was not found")
	}
	if prompt.Description != "根据主角和场景生成故事开头" {
		t.Errorf("expected description mismatch: got %s", prompt.Description)
	}
	if len(prompt.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(prompt.Arguments))
	}

	ctx := context.Background()
	req := &GetPromptRequest{
		Params: struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments,omitempty"`
		}{
			Name: "story_generator",
			Arguments: map[string]string{
				"hero":  "猫",
				"scene": "森林",
			},
		},
	}

	result, err := manager.prompts["story_generator"].Handler(ctx, req)
	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Content.(TextContent).Text != "主角是：猫，场景是：森林" {
		t.Errorf("unexpected prompt message content: %s", result.Messages[0].Content.(TextContent).Text)
	}

}

func TestHandleGetPrompt_DefaultHandler(t *testing.T) {
	manager := newPromptManager()
	prompt := &Prompt{
		Name:        "default_prompt",
		Description: "这是一个没有自定义 handler 的提示",
		Arguments: []PromptArgument{
			{Name: "keyword", Description: "关键词", Required: true},
		},
	}
	manager.registerPrompt(prompt, nil)

	// 构造 JSONRPCRequest
	req := &JSONRPCRequest{
		ID:      1,
		JSONRPC: "2.0",
		Params: map[string]interface{}{
			"name": "default_prompt",
			"arguments": map[string]interface{}{
				"keyword": "测试",
			},
		},
	}

	resp, err := manager.handleGetPrompt(context.Background(), req)
	if err != nil {
		t.Fatalf("handleGetPrompt error: %v", err)
	}

	result, ok := resp.(*GetPromptResult)
	if !ok {
		t.Fatalf("unexpected response type: %T", resp)
	}
	if len(result.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Content.(TextContent).Text == "" {
		t.Errorf("expected text content, got empty string")
	}
}

func TestParseGetPromptParams_Invalid(t *testing.T) {
	// 缺少 name 字段
	req := &JSONRPCRequest{
		ID:     2,
		Params: map[string]interface{}{},
	}
	_, _, errResp, ok := parseGetPromptParams(req)
	if ok {
		t.Fatal("expected parse failure but got success")
	}
	if errResp == nil {
		t.Fatal("expected error response")
	}
}

func TestHandleListPrompts(t *testing.T) {
	manager := newPromptManager()
	manager.registerPrompt(&Prompt{
		Name:        "p1",
		Description: "desc1",
	}, nil)
	manager.registerPrompt(&Prompt{
		Name:        "p2",
		Description: "desc2",
	}, nil)

	resp, err := manager.handleListPrompts(context.Background(), &JSONRPCRequest{})
	if err != nil {
		t.Fatalf("handleListPrompts error: %v", err)
	}
	result, ok := resp.(*ListPromptsResult)
	if !ok {
		t.Fatalf("expected ListPromptsResult, got %T", resp)
	}
	if len(result.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(result.Prompts))
	}
}

func TestRegisterPrompts_Batch(t *testing.T) {
	manager := newPromptManager()
	prompts := []*Prompt{
		{Name: "batch_1"},
		{Name: "batch_2"},
	}
	manager.registerPrompts(prompts, nil)

	for _, p := range prompts {
		if _, exists := manager.getPrompt(p.Name); !exists {
			t.Errorf("prompt %s not registered", p.Name)
		}
	}
}

func TestHandleCompletionComplete_InvalidRef(t *testing.T) {
	manager := newPromptManager()
	req := &JSONRPCRequest{
		ID:     3,
		Params: map[string]interface{}{"ref": map[string]interface{}{"type": "unknown"}},
	}
	resp, err := manager.handleCompletionComplete(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected JSONRPC error response")
	}
}

func TestBuildPromptMessages_MissingRequired(t *testing.T) {
	prompt := &Prompt{
		Name: "incomplete",
		Arguments: []PromptArgument{
			{Name: "field1", Required: true},
			{Name: "field2", Required: false},
		},
	}
	args := map[string]interface{}{}
	messages := buildPromptMessages(prompt, args)
	if len(messages) != 1 {
		t.Fatal("expected 1 message")
	}
	text := messages[0].Content.(TextContent).Text
	if !strings.Contains(text, "[not provided]") {
		t.Error("expected '[not provided]' for missing required parameter")
	}
}

//{
//  "jsonrpc": "2.0",
//  "id": 1,
//  "method": "getPrompt",
//  "params": {
//    "name": "story_generator",
//    "arguments": {
//      "hero": "猫",
//      "scene": "森林"
//    }
//  }
//}
//{
//  "description": "根据主角和场景生成故事开头",
//  "messages": [
//    {
//      "role": "user",
//      "content": {
//        "type": "text",
//        "text": "This is an example rendering of the story_generator prompt.\nParameter hero: 猫\nParameter scene: 森林"
//      }
//    }
//  ]
//}
