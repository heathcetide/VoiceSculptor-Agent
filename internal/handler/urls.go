package handlers

import (
	voiceSculptor "VoiceSculptor"
	"VoiceSculptor/internal/apidocs"
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/config"
	"VoiceSculptor/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	db *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
	}
}

func (h *Handlers) Register(engine *gin.Engine) {
	r := engine.Group(config.GlobalConfig.APIPrefix)
	// 注册系统模块功能
	h.registerSystemRoutes(r)

	// 注册各模块的路由
	h.registerAuthRoutes(r)
	h.registerAssistantRoutes(r)
	h.registerChatRoutes(r)

	objs := h.GetObjs()
	voiceSculptor.RegisterObjects(r, objs)
	if config.GlobalConfig.DocsPrefix != "" {
		var objDocs []apidocs.WebObjectDoc
		for _, obj := range objs {
			objDocs = append(objDocs, apidocs.GetWebObjectDocDefine(config.GlobalConfig.APIPrefix, obj))
		}
		apidocs.RegisterHandler(config.GlobalConfig.DocsPrefix, engine, h.GetDocs(), objDocs, h.db)
	}
	if config.GlobalConfig.AdminPrefix != "" {
		admin := r.Group(config.GlobalConfig.AdminPrefix)
		h.RegisterAdmin(admin)
	}
}

// User Module
func (h *Handlers) registerAuthRoutes(r *gin.RouterGroup) {
	r.Use(middleware.InjectDB(h.db))
	auth := r.Group(config.GlobalConfig.AuthPrefix)
	{
		// register
		auth.GET("register", h.handleUserSignupPage)

		auth.POST("register", h.handleUserSignup)

		auth.POST("send/email", h.handleSendEmailCode)

		// login
		auth.GET("login", h.handleUserSigninPage)

		auth.POST("login", h.handleUserSignin)

		// logout
		auth.GET("logout", h.handleUserLogout)

		auth.GET("info", h.handleUserInfo)

		auth.GET("reset-password", h.handleUserResetPasswordPage)
	}
}

func (h *Handlers) registerAssistantRoutes(r *gin.RouterGroup) {
	assistant := r.Group("assistant")
	{
		assistant.POST("", h.CreateAssistant)

		assistant.GET("", h.ListAssistants)

		assistant.GET("/:id", h.GetAssistant)

		assistant.PUT("/:id", h.UpdateAssistant)

		assistant.DELETE("/:id", h.DeleteAssistant)

	}
}

func (h *Handlers) registerChatRoutes(r *gin.RouterGroup) {
	chat := r.Group("chat")
	{
		chat.POST("/start", h.Chat)

		chat.POST("/stop", h.StopChat)
	}
}

func (h *Handlers) registerSystemRoutes(r *gin.RouterGroup) {
	system := r.Group("/system")
	{
		system.POST("/rate-limiter/config", h.UpdateRateLimiterConfig)

		system.GET("/health", h.HealthCheck)
	}
}

func (h *Handlers) GetObjs() []voiceSculptor.WebObject {
	return []voiceSculptor.WebObject{
		{
			Group:       "VoiceSculptor",
			Desc:        "用户",
			Model:       models.User{},
			Name:        "user",
			Filterables: []string{"UpdateAt", "CreatedAt"},
			Editables:   []string{"Email", "Phone", "FirstName", "LastName", "DisplayName", "IsSuperUser", "Enabled"},
			Searchables: []string{},
			Orderables:  []string{"UpdatedAt"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				if isCreate {
					return h.db
				}
				return h.db.Where("deleted_at", nil)
			},
			BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
				return nil
			},
		},
	}
}

func (h *Handlers) RegisterAdmin(router *gin.RouterGroup) {
	adminObjs := models.GetHibiscusAdminObjects()
	iconAssistant, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_assistant.svg")
	iconChatLog, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_chat_log.svg")
	iconUserCredential, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_user_credential.svg")
	admins := []models.AdminObject{
		{
			Model:       &models.Assistant{},
			Group:       "Business",
			Name:        "Assistant",
			Desc:        "This is a definition of AI assistant, including the use of prompts and so on.",
			Shows:       []string{"ID", "Name", "SystemPrompt", "Instruction", "PersonaTag", "MaxTokens", "Temperature", "CreatedAt"},
			Editables:   []string{"ID", "Name", "SystemPrompt", "Instruction", "PersonaTag", "MaxTokens", "Temperature", "CreatedAt"},
			Orderables:  []string{"UpdatedAt"},
			Searchables: []string{"Name"},
			Requireds:   []string{"Name"},
			Icon:        &models.AdminIcon{SVG: string(iconAssistant)},
		},
		{
			Model:       &models.ChatLog{},
			Group:       "Business",
			Name:        "ChatLog",
			Desc:        "This is a conversation log, which records the AI conversation log.",
			Shows:       []string{"ID", "Input", "Output", "PromptTokens", "OutputTokens", "CreatedAt"},
			Editables:   []string{"ID", "Input", "Output", "PromptTokens", "OutputTokens", "UpdatedAt"},
			Orderables:  []string{"UpdatedAt"},
			Searchables: []string{"UserID"},
			Requireds:   []string{"UserID"},
			Icon:        &models.AdminIcon{SVG: string(iconChatLog)},
		},
		{
			Model:       &models.UserCredential{},
			Group:       "Business",
			Name:        "UserCredential",
			Desc:        "This is a user credential used to define which user resources.",
			Shows:       []string{"ID", "Name", "LLMProvider", "LLMApiKey", "LLMApiURL", "Quota", "Used"},
			Editables:   []string{"ID", "Name", "LLMProvider", "LLMApiKey", "LLMApiURL", "Quota", "Used"},
			Orderables:  []string{"UpdatedAt"},
			Searchables: []string{"LLMProvider"},
			Requireds:   []string{"LLMProvider"},
			Icon:        &models.AdminIcon{SVG: string(iconUserCredential)},
		},
	}
	models.RegisterAdmins(router, h.db, append(adminObjs, admins...))
}
