// model.go
package client

// PBXMessage 表示一个呼叫请求消息
type PBXMessage struct {
	Command string       `json:"command"` // 操作类型，如 'invite', 'tts'
	Option  *CallOptions `json:"option,omitempty"`
	Text    string       `json:"text,omitempty"`
	PlayId  string       `json:"playId,omitempty"`
}

// CallOptions 包含呼叫配置的详细信息
type CallOptions struct {
	Asr   *AsrConfig `json:"asr,omitempty"`   // 自动语音识别配置
	Tts   *TtsConfig `json:"tts,omitempty"`   // 语音合成配置
	Offer string     `json:"offer,omitempty"` // SDP Offer 信息
}

// AsrConfig 自动语音识别配置
type AsrConfig struct {
	Provider  string `json:"provider"`
	AppId     string `json:"appId"`
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	Language  string `json:"language"`
}

// TtsConfig 文本转语音配置
type TtsConfig struct {
	Provider  string  `json:"provider"`
	Speaker   string  `json:"speaker"`
	AppId     string  `json:"appId"`
	SecretId  string  `json:"secretId"`
	SecretKey string  `json:"secretKey"`
	Speed     float32 `json:"speed"`
	Volume    int     `json:"volume"`
}

// EventMessage 表示服务端发送的事件通知
type EventMessage struct {
	Event     string                 `json:"event"`
	TrackId   string                 `json:"trackId,omitempty"`
	Timestamp *uint64                `json:"timestamp,omitempty"`
	Key       string                 `json:"key,omitempty"`
	Duration  uint32                 `json:"duration,omitempty"`
	SDP       string                 `json:"sdp,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Text      string                 `json:"text,omitempty"`
}
