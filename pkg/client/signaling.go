// signaling.go
package client

import (
	"VoiceSculptor/pkg/llm"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
)

type SignalingClient struct {
	conn      *websocket.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	media     *MediaHandler
	llmClient *llm.OpenAIClient
	logger    *logrus.Logger
	promptCfg PromptConfig
	recvDone  chan struct{}
}

func NewSignalingClient(ctx context.Context, serverAddr string, initial PBXMessage, media *MediaHandler, llmClient *llm.OpenAIClient, logger *logrus.Logger, prompt PromptConfig) (*SignalingClient, error) {
	ctx, cancel := context.WithCancel(ctx)

	conn, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	msg, err := json.Marshal(initial)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to marshal initial PBXMessage: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to send initial message: %w", err)
	}
	log.Println("Send invite command to RustPBX....")

	client := &SignalingClient{
		conn:      conn,
		ctx:       ctx,
		cancel:    cancel,
		media:     media,
		llmClient: llmClient,
		logger:    logger,
		promptCfg: prompt,
		recvDone:  make(chan struct{}),
	}
	go client.listen()
	return client, nil
}

func (s *SignalingClient) listen() {
	defer close(s.recvDone)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			typeVal, data, err := s.conn.ReadMessage()
			if err != nil {
				s.logger.Errorf("read error: %v", err)
				return
			}
			if typeVal != websocket.TextMessage {
				continue
			}
			var evt EventMessage
			if err := json.Unmarshal(data, &evt); err != nil {
				s.logger.Warnf("invalid message: %s", data)
				continue
			}
			s.handleEvent(evt)
		}
	}
}

func (s *SignalingClient) handleEvent(evt EventMessage) {
	s.logger.Infof("Received event: %s", evt.Event)
	switch evt.Event {
	case "answer":
		s.logger.Infof("Received answer: %s", evt.SDP)
		err := s.media.SetupAnswer(evt.SDP)
		if err != nil {
			s.logger.Errorf("failed to apply answer: %v", err)
		}
	case "asrFinal":
		go s.handleASRFinal(evt.Text)
	default:
		s.logger.Infof("Unhandled event: %s", evt.Event)
	}
}

func (s *SignalingClient) handleASRFinal(input string) {
	s.logger.Infof("Constructed prompt: %s   -    %s     -    %s", s.promptCfg.SystemPrompt, input, s.promptCfg.Instruction)

	// 推送用户输入给 SSE
	if s.llmClient != nil && s.llmClient.SSE != nil {
		s.llmClient.SSE <- "[user] " + input
	}

	reply, err := s.llmClient.GenerateText(s.promptCfg.SystemPrompt, input, s.promptCfg.Instruction, s.promptCfg.PersonaTag)
	if err != nil {
		s.logger.Errorf("LLM generation failed: %v", err)
		return
	}

	resp := PBXMessage{
		Command: "tts",
		Text:    reply,
	}
	msg, err := json.Marshal(resp)
	if err != nil {
		s.logger.Errorf("marshal tts message failed: %v", err)
		return
	}
	if err := s.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		s.logger.Errorf("send tts message failed: %v", err)
	}
}

func (s *SignalingClient) Close() {
	s.cancel()
	_ = s.conn.Close()
	<-s.recvDone
}
