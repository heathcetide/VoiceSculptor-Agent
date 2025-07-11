package client

import (
	"VoiceSculptor/pkg/llm"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"sync"
	"time"
)

// Config holds the configuration for initializing the SDK
type Config struct {
	ICEURL     string
	ServerAddr string
	ASR        AsrConfig
	TTS        TtsConfig
	OpenAIKey  string
	Logger     *logrus.Logger // Optional
	Prompt     PromptConfig   `json:"prompt"`
}

type PromptConfig struct {
	SystemPrompt   string  `json:"systemPrompt"`   // 系统角色
	Instruction    string  `json:"instruction"`    // 具体引导语
	PersonaTag     string  `json:"personaTag"`     // 角色标识
	Temperature    float32 `json:"temperature"`    // 发散度
	MaxTokens      int     `json:"maxTokens"`      // 最大响应长度
	HistoryEnabled bool    `json:"historyEnabled"` // 是否启用上下文记忆
}

// SDK is the main exposed struct for users
type Client struct {
	ctx          context.Context
	cancel       context.CancelFunc
	media        *MediaHandler
	signal       *SignalingClient
	llmClient    *llm.OpenAIClient
	logger       *logrus.Logger
	config       Config
	started      bool
	startedMutex sync.Mutex
	sseChan      chan string
}

// NewSDK initializes the SDK with the given config
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	ctx, cancel := context.WithCancel(ctx)

	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.InfoLevel)
	}

	mediaHandler, err := NewMediaHandler(ctx, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create media handler: %w", err)
	}

	llmClient := llm.NewOpenAIClient(cfg.OpenAIKey)

	sdk := &Client{
		ctx:       ctx,
		cancel:    cancel,
		media:     mediaHandler,
		logger:    logger,
		config:    cfg,
		llmClient: llmClient,
		sseChan:   make(chan string, 100),
	}

	go func() {
		for msg := range llmClient.SSE {
			sdk.sseChan <- msg
		}
	}()

	return sdk, nil
}

// StartCall initializes media, sets up signaling, and sends invite
func (s *Client) StartCall() error {
	s.startedMutex.Lock()
	defer s.startedMutex.Unlock()
	if s.started {
		return errors.New("call already started")
	}

	iceServers, err := fetchICEServers(s.config.ICEURL)
	if err != nil {
		return fmt.Errorf("failed to fetch ICE servers: %w", err)
	}

	offer, err := s.media.Setup("g722", iceServers)
	if err != nil {
		return fmt.Errorf("failed to setup media: %w", err)
	}

	pbx := PBXMessage{
		Command: "invite",
		Option: &CallOptions{
			Asr:   &s.config.ASR,
			Tts:   &s.config.TTS,
			Offer: offer,
		},
	}

	s.signal, err = NewSignalingClient(s.ctx, s.config.ServerAddr, pbx, s.media, s.llmClient, s.logger, s.config.Prompt)
	if err != nil {
		return fmt.Errorf("failed to setup signaling: %w", err)
	}

	s.started = true
	// Start a goroutine to handle voice input timeout and automatic disconnect
	go s.handleVoiceTimeout()
	return nil
}

// 在handleVoiceTimeout中使用llmClient向用户发送消息
func (s *Client) handleVoiceTimeout() {
	voiceTimer := time.NewTimer(20 * time.Second)
	disconnectTimer := time.NewTimer(80 * time.Second)

	for {
		select {
		case <-voiceTimer.C:
			// 20秒未收到语音输入，询问是否在线
			s.llmClient.SendMessage("请问您还在线吗？")
			voiceTimer.Reset(20 * time.Second) // 重置计时器

		case <-disconnectTimer.C:
			// 60秒未收到语音输入，自动断开连接
			s.Close()
			s.logger.Info("未收到语音输入，自动断开连接")
			return
		}
	}
}

// Close gracefully shuts down the SDK
func (s *Client) Close() {
	s.cancel()
	if s.media != nil {
		_ = s.media.Stop()
	}
	if s.signal != nil {
		s.signal.Close()
	}
}

// fetchICEServers fetches ICE servers from the provided URL
func fetchICEServers(url string) ([]webrtc.ICEServer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var iceServers []webrtc.ICEServer
	if err := json.NewDecoder(resp.Body).Decode(&iceServers); err != nil {
		return nil, err
	}
	return iceServers, nil
}

// SSEChannel 返回 SSE 消息通道（只读）
func (c *Client) SSEChannel() <-chan string {
	return c.sseChan
}
