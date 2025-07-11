package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

const (
	GroupRoleAdmin  = "admin"
	GroupRoleMember = "member"
	SigInitDBConfig = "system.init"
)

type BaseModel struct {
	ID       uint `gorm:"primaryKey"`
	CreateAt time.Time
	UpdateAt time.Time
	CreateBy string
	UpdateBy string
	Version  int16
	isDel    int8 `gorm:"index"`
}

type User struct {
	ID        uint       `json:"-" gorm:"primaryKey"`
	CreatedAt time.Time  `json:"-" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"-" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Email              string `json:"email" gorm:"size:128;uniqueIndex"`
	EmailNotifications bool   `json:"emailNotifications"`

	Password    string     `json:"-" gorm:"size:128"`
	Phone       string     `json:"phone,omitempty" gorm:"size:64;index"`
	FirstName   string     `json:"firstName,omitempty" gorm:"size:128"`
	LastName    string     `json:"lastName,omitempty" gorm:"size:128"`
	DisplayName string     `json:"displayName,omitempty" gorm:"size:128"`
	IsSuperUser bool       `json:"-"`
	IsStaff     bool       `json:"isStaff,omitempty"`
	Enabled     bool       `json:"-"`
	Activated   bool       `json:"-"`
	LastLogin   *time.Time `json:"lastLogin,omitempty"`
	LastLoginIP string     `json:"-" gorm:"size:128"`

	Source    string `json:"-" gorm:"size:64;index"`
	Locale    string `json:"locale,omitempty" gorm:"size:20"`
	Timezone  string `json:"timezone,omitempty" gorm:"size:200"`
	AuthToken string `json:"token,omitempty" gorm:"-"`

	Avatar       string `json:"avatar,omitempty"`
	Gender       string `json:"gender,omitempty"`
	City         string `json:"city,omitempty"`
	Region       string `json:"region,omitempty"`
	Country      string `json:"country,omitempty"`
	Extra        string `json:"extra,omitempty"`
	PrivateExtra string `json:"privateExtra,omitempty"`
}

type UserCredential struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index;"`               // 关联到用户
	Name      string `json:"name"`                 // 应用名称 or 用途备注
	APIKey    string `gorm:"uniqueIndex;not null"` // 用于认证
	APISecret string `gorm:"not null"`             // 用于签名校验

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

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type GroupPermission struct {
	Permissions []string
}

type Group struct {
	ID         uint            `json:"-" gorm:"primaryKey"`
	CreatedAt  time.Time       `json:"-" gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `json:"-"`
	Name       string          `json:"name" gorm:"size:200"`
	Type       string          `json:"type" gorm:"size:24;index"`
	Extra      string          `json:"extra"`
	Permission GroupPermission `json:"-"`
}

// 实现 driver.Valuer 接口
func (gp GroupPermission) Value() (driver.Value, error) {
	return json.Marshal(gp)
}

// 实现 sql.Scanner 接口
func (gp *GroupPermission) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert value to []byte")
	}
	return json.Unmarshal(bytes, gp)
}

type GroupMember struct {
	ID        uint      `json:"-" gorm:"primaryKey"`
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime"`
	UserID    uint      `json:"-"`
	User      User      `json:"user"`
	GroupID   uint      `json:"-"`
	Group     Group     `json:"group"`
	Role      string    `json:"role" gorm:"size:60"`
}
