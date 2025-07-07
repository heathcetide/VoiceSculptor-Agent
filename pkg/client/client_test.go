package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSDK_StartCall(t *testing.T) {
	cfg := Config{
		ICEURL:     "http://localhost:8080/iceservers",
		ServerAddr: "ws://localhost:8080/call/webrtc",
		ASR: AsrConfig{
			Provider:  "tencent",
			AppId:     "1325039295",
			SecretId:  "AKIDb4KNEWpvvx23yqdFh8Xlq9SeptmWadju",
			SecretKey: "Khx9wfaTNYiP5fFl7XsDYmqxwhLrfP1U",
			Language:  "zh-cn",
		},
		TTS: TtsConfig{
			Provider:  "tencent",
			Speaker:   "101016", // 301030
			AppId:     "1325039295",
			SecretId:  "AKIDb4KNEWpvvx23yqdFh8Xlq9SeptmWadju",
			SecretKey: "Khx9wfaTNYiP5fFl7XsDYmqxwhLrfP1U",
			Speed:     1.0,
			Volume:    5,
		},
		OpenAIKey: "83fb4b9faddc98a274664b5bd4141aa7.6nNum7w223OgxqV3",
		Prompt: PromptConfig{
			SystemPrompt:   `你是一位优雅而富有耐心的编程导师，擅长用温柔且专业的语气指导程序员解决技术问题，并鼓励他们保持信心。`,
			Instruction:    `请使用温和、鼓励的语气，尽量用简洁易懂的语言回答问题。避免批评，强调理解与支持。`,
			PersonaTag:     "elegant_programmer_mentor",
			Temperature:    0.6, // 稳重风格
			MaxTokens:      150, // 限制字数保持对话流畅
			HistoryEnabled: false,
		},
	}

	sdk, err := NewClient(context.Background(), cfg)
	assert.NoError(t, err)
	defer sdk.Close()

	err = sdk.StartCall()
	assert.NoError(t, err)

	t.Log("Waiting for connection...")
	time.Sleep(10 * time.Second)

	// Nothing to assert on actual media yet, but confirms no crash
	t.Log("Call initiated successfully")
}
