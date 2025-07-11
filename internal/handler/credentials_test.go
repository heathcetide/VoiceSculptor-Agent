package handlers

import (
	"VoiceSculptor/internal/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 初始化测试用的路由和内存数据库
func setupTestCredentialRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移模型
	err = db.AutoMigrate(&models.User{}, &models.UserCredential{})
	assert.NoError(t, err)

	// 插入测试用户
	err = db.Create(mockUser).Error
	assert.NoError(t, err)

	router := gin.New()
	router.Use(mockAuthMiddleware())

	h := &Handlers{db: db}

	// 注册凭证相关的路由
	api := router.Group("/api")
	{
		api.POST("/credential", h.handleCreateCredential)
		api.GET("/credentials", h.handleGetCredential)
	}

	return router, db
}

//
//// TestHandleCreateCredential_Success 测试创建凭证成功的情况
//func TestHandleCreateCredential_Success(t *testing.T) {
//	r, _ := setupTestCredentialRouter(t)
//
//	// 构造请求体
//	credential := models.UserCredentialRequest{
//		Name:         "Test Credential",
//		APIKey:       "apikey123",
//		APISecret:    "apisecret123",
//		AsrProvider:  "asr_provider",
//		AsrAppID:     "asr_appid",
//		AsrSecretID:  "asr_secretid",
//		AsrSecretKey: "asr_secretkey",
//		AsrLanguage:  "zh-CN",
//	}
//
//	body, _ := json.Marshal(credential)
//	req := httptest.NewRequest("POST", "/api/credential", bytes.NewBuffer(body))
//	req.Header.Set("Content-Type", "application/json")
//	resp := httptest.NewRecorder()
//
//	r.ServeHTTP(resp, req)
//
//	assert.Equal(t, http.StatusOK, resp.Code)
//	var res map[string]interface{}
//	err := json.Unmarshal(resp.Body.Bytes(), &res)
//	assert.NoError(t, err)
//	assert.True(t, res["success"].(bool))
//}

//// TestHandleCreateCredential_InvalidJSON 测试无效 JSON 输入
//func TestHandleCreateCredential_InvalidJSON(t *testing.T) {
//	r, _ := setupTestCredentialRouter(t)
//
//	req := httptest.NewRequest("POST", "/api/credential", bytes.NewBuffer([]byte("invalid-json")))
//	req.Header.Set("Content-Type", "application/json")
//	resp := httptest.NewRecorder()
//
//	r.ServeHTTP(resp, req)
//
//	assert.Equal(t, http.StatusBadRequest, resp.Code)
//	var res map[string]interface{}
//	err := json.Unmarshal(resp.Body.Bytes(), &res)
//	assert.NoError(t, err)
//	assert.False(t, res["success"].(bool))
//}
//
//// TestHandleGetCredential_Success 测试获取用户所有凭证信息
//func TestHandleGetCredential_Success(t *testing.T) {
//	r, db := setupTestCredentialRouter(t)
//
//	// 手动插入一条凭证记录
//	cred := &models.UserCredential{
//		UserID:       mockUser.ID,
//		APIKey:       "apikey123",
//		APISecret:    "apisecret123",
//		AsrProvider:  "asr_provider",
//		AsrAppID:     "asr_appid",
//		AsrSecretID:  "asr_secretid",
//		AsrSecretKey: "asr_secretkey",
//		AsrLanguage:  "zh-CN",
//		LLMProvider:  "llm_provider",
//		LLMApiKey:    "llm_apikey",
//		LLMApiURL:    "llm_apiurl",
//	}
//
//	db.Create(cred)
//
//	req := httptest.NewRequest("GET", "/api/credentials", nil)
//	resp := httptest.NewRecorder()
//	r.ServeHTTP(resp, req)
//
//	assert.Equal(t, http.StatusOK, resp.Code)
//	var res map[string]interface{}
//	err := json.Unmarshal(resp.Body.Bytes(), &res)
//	assert.NoError(t, err)
//	assert.True(t, res["success"].(bool))
//
//	// 验证返回的数据是否包含我们插入的凭证
//	data := res["data"].([]interface{})
//	assert.Len(t, data, 1)
//	firstCred := data[0].(map[string]interface{})
//	assert.Equal(t, cred.APIKey, firstCred["apiKey"])
//	assert.Equal(t, cred.AsrProvider, firstCred["asrProvider"])
//}
