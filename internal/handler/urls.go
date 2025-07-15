package handlers

import (
	voiceSculptor "VoiceSculptor"
	"VoiceSculptor/internal/apidocs"
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/config"
	"VoiceSculptor/pkg/middleware"
	"VoiceSculptor/pkg/notification"
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

	// Register Global Singleton DB
	r.Use(middleware.InjectDB(h.db))
	// Register System Module Routes
	h.registerSystemRoutes(r)

	// Register Business Module Routes
	h.registerAuthRoutes(r)
	h.registerAssistantRoutes(r)
	h.registerChatRoutes(r)
	h.registerNotificationRoutes(r)
	h.registerCredentialsRoutes(r)
	h.registerGroupRoutes(r)

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
	auth := r.Group(config.GlobalConfig.AuthPrefix)
	{
		// register
		auth.GET("/register", h.handleUserSignupPage)

		auth.POST("/register", h.handleUserSignup)

		auth.POST("/register/email", h.handleUserSignupByEmail)

		auth.POST("/send/email", h.handleSendEmailCode)

		// login
		auth.GET("/login", h.handleUserSigninPage)

		auth.POST("/login", h.handleUserSignin)

		auth.POST("/login/email", h.handleUserSigninByEmail)

		// logout
		auth.GET("/logout", models.AuthRequired, h.handleUserLogout)

		auth.GET("/info", models.AuthRequired, h.handleUserInfo)

		auth.GET("/reset-password", h.handleUserResetPasswordPage)

		// update
		auth.PUT("/update", models.AuthRequired, h.handleUserUpdate)

		auth.PUT("/update/preferences", models.AuthRequired, h.handleUserUpdatePreferences)
	}
}

func (h *Handlers) registerAssistantRoutes(r *gin.RouterGroup) {
	assistant := r.Group("assistant")
	{
		assistant.POST("add", models.AuthRequired, h.CreateAssistant)

		assistant.GET("", models.AuthRequired, h.ListAssistants)

		assistant.GET("/:id", models.AuthRequired, h.GetAssistant)

		assistant.PUT("/:id", models.AuthRequired, h.UpdateAssistant)

		assistant.DELETE("/:id", models.AuthRequired, h.DeleteAssistant)

		assistant.GET("/voiceSculptor/client/:id/loader.js", h.ServeVoiceSculptorLoaderJS)
	}
}

func (h *Handlers) registerChatRoutes(r *gin.RouterGroup) {
	chat := r.Group("chat")
	chat.Use(models.AuthApiRequired)
	{
		chat.POST("start", h.Chat)

		chat.POST("stop", h.StopChat)

		chat.GET("stream", h.ChatStream)

		chat.GET("chat-session-log", h.getChatSessionLog)

		chat.GET("chat-session-log/:id", h.getChatSessionLogDetail)
	}
}

func (h *Handlers) registerNotificationRoutes(r *gin.RouterGroup) {
	notificationGroup := r.Group("notification")
	{
		notificationGroup.GET("unread-count", models.AuthRequired, h.handleUnReadNotificationCount)

		notificationGroup.GET("", models.AuthRequired, h.handleListNotifications)

		notificationGroup.POST("readAll", models.AuthRequired, h.handleAllNotifications)

		notificationGroup.PUT("/read/:id", models.AuthRequired, h.handleMarkNotificationAsRead)

		notificationGroup.DELETE("/:id", models.AuthRequired, h.handleDeleteNotification)
	}
}

func (h *Handlers) registerSystemRoutes(r *gin.RouterGroup) {
	system := r.Group("system")
	{
		system.POST("/rate-limiter/config", h.UpdateRateLimiterConfig)

		system.GET("/health", h.HealthCheck)
	}
}

func (h *Handlers) registerCredentialsRoutes(r *gin.RouterGroup) {
	credential := r.Group("credentials")
	{
		credential.POST("/", models.AuthRequired, h.handleCreateCredential)

		credential.GET("/", models.AuthRequired, h.handleGetCredential)
	}
}

func (h *Handlers) registerGroupRoutes(r *gin.RouterGroup) {
	group := r.Group("group")
	group.OPTIONS("/*cors", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.AbortWithStatus(204)
	})
	group.Use(models.AuthRequired)
	{
		group.POST("/", h.CreateGroup)

		group.GET("/", h.ListGroups)

		group.GET("/:id", h.GetGroup)

		group.PUT("/:id", h.UpdateGroup)

		group.DELETE("/:id", h.DeleteGroup)
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
	iconInternalNotification, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_internal_notification.svg")
	iconPrompt, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_prompt_model.svg")
	iconPromptArg, _ := voiceSculptor.EmbedStaticAssets.ReadFile("static/img/icon_prompt_args.svg")
	admins := []models.AdminObject{
		{
			Model:       &models.Assistant{},
			Group:       "Business",
			Name:        "Assistant",
			Desc:        "This is a definition of AI assistant, including the use of prompts and so on.",
			Shows:       []string{"ID", "Name", "SystemPrompt", "Instruction", "PersonaTag", "MaxTokens", "Temperature", "JsSourceID", "CreatedAt"},
			Editables:   []string{"ID", "Name", "SystemPrompt", "Instruction", "PersonaTag", "MaxTokens", "Temperature", "JsSourceID", "CreatedAt"},
			Orderables:  []string{"UpdatedAt"},
			Searchables: []string{"Name"},
			Requireds:   []string{"Name"},
			Icon:        &models.AdminIcon{SVG: string(iconAssistant)},
		},
		{
			Model:       &models.ChatSessionLog{},
			Group:       "Business",
			Name:        "ChatSessionLog",
			Desc:        "This is a conversation log, which records the AI conversation log.",
			Shows:       []string{"ID", "SessionID", "Content", "CreatedAt", "UserID"},
			Editables:   []string{"ID", "SessionID", "Content", "CreatedAt", "UserID"},
			Orderables:  []string{"CreatedAt"},
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
		{
			Model:       &notification.InternalNotification{},
			Group:       "Business",
			Name:        "InternalNotification",
			Desc:        "This is a notification used to notify the user of the system.",
			Shows:       []string{"ID", "Title", "Read", "CreatedAt"},
			Editables:   []string{"ID", "UserID", "Title", "Content", "Read", "CreatedAt"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Title"},
			Icon:        &models.AdminIcon{SVG: string(iconInternalNotification)},
		},
		{
			Model:       &models.PromptModel{},
			Group:       "Business",
			Name:        "PromptModel",
			Desc:        "This is a PromptModel, can quick build prompt",
			Shows:       []string{"ID", "Name", "Description", "CreatedAt", "UpdatedAt"},
			Editables:   []string{"ID", "Name", "Description", "CreatedAt", "UpdatedAt"},
			Orderables:  []string{"CreatedAt"},
			Searchables: []string{"Name"},
			Icon:        &models.AdminIcon{SVG: string(iconPrompt)},
		},
		{
			Model:       &models.PromptArgModel{},
			Group:       "Business",
			Name:        "PromptArgModel",
			Desc:        "This is a PromptModel Args to fill model",
			Shows:       []string{"ID", "Name", "Description", "Required", "PromptID"},
			Editables:   []string{"ID", "Name", "Description", "Required", "PromptID"},
			Orderables:  []string{"ID"},
			Searchables: []string{"Name"},
			Icon:        &models.AdminIcon{SVG: string(iconPromptArg)},
		},
	}
	models.RegisterAdmins(router, h.db, append(adminObjs, admins...))
}
