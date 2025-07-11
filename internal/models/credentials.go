package models

import (
	"VoiceSculptor/pkg/util"
	"errors"
	"gorm.io/gorm"
)

type UserCredentialRequest struct {
	Name string `json:"name"` // 应用名称 or 用途备注

	LLMProvider string `json:"llmProvider"`
	LLMApiKey   string `json:"llmApiKey"`
	LLMApiURL   string `json:"llmApiUrl"`

	AsrProvider  string `json:"asrProvider"`
	AsrAppID     string `json:"asrAppId"`
	AsrSecretID  string `json:"asrSecretId"`
	AsrSecretKey string `json:"asrSecretKey"`
	AsrLanguage  string `json:"language"`

	TtsProvider  string `json:"ttsProvider"`
	TTSAppID     string `json:"ttsAppId"`
	TTSSecretID  string `json:"ttsSecretId"`
	TTSSecretKey string `json:"ttsSecretKey"`
}

// CreateUserCredential 创建用户凭证
func CreateUserCredential(db *gorm.DB, userID uint, credential *UserCredentialRequest) (*UserCredential, error) {
	apiKey, err := util.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	apiSecret, err := util.GenerateSecureToken(64)
	if err != nil {
		return nil, err
	}

	userCred := &UserCredential{
		UserID:       userID,
		APIKey:       apiKey,
		APISecret:    apiSecret,
		Name:         credential.Name,
		LLMProvider:  credential.LLMProvider,
		LLMApiKey:    credential.LLMApiKey,
		LLMApiURL:    credential.LLMApiURL,
		AsrProvider:  credential.AsrProvider,
		AsrAppID:     credential.AsrAppID,
		AsrSecretID:  credential.AsrSecretID,
		AsrSecretKey: credential.AsrSecretKey,
		AsrLanguage:  credential.AsrLanguage,
		TtsProvider:  credential.TtsProvider,
		TTSAppID:     credential.TTSAppID,
		TTSSecretID:  credential.TTSSecretID,
		TTSSecretKey: credential.TTSSecretKey,
	}

	err = db.Create(userCred).Error
	if err != nil {
		return nil, err
	}

	return userCred, nil
}

// GetUserCredentials 根据用户ID获取其所有的凭证信息
func GetUserCredentials(db *gorm.DB, userID uint) ([]*UserCredential, error) {
	var credentials []*UserCredential
	err := db.Where("user_id = ?", userID).Find(&credentials).Error
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func GetUserCredentialByApiSecretAndApiKey(db *gorm.DB, apiKey, apiSecret string) (*UserCredential, error) {
	var credential UserCredential
	result := db.Where("api_key = ? AND api_secret = ?", apiKey, apiSecret).First(&credential)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &credential, nil
}
