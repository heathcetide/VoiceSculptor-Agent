package main

import (
	voiceSculptor "VoiceSculptor"
	handlers "VoiceSculptor/internal/handler"
	"VoiceSculptor/internal/listeners"
	"VoiceSculptor/internal/models"
	"VoiceSculptor/internal/task"
	"VoiceSculptor/pkg/config"
	constants "VoiceSculptor/pkg/constant"
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/middleware"
	"VoiceSculptor/pkg/notification"
	"VoiceSculptor/pkg/prompt"
	"VoiceSculptor/pkg/util"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"time"
)

type VoiceSculptorApp struct {
	db       *gorm.DB
	handlers *handlers.Handlers
}

func NewVoiceSculptorApp(db *gorm.DB) *VoiceSculptorApp {
	return &VoiceSculptorApp{
		db:       db,
		handlers: handlers.NewHandlers(db),
	}
}

func initDefaultConfigs(db *gorm.DB) error {
	defaults := []util.Config{
		{Key: constants.KEY_SITE_URL, Desc: "站点网址", Autoload: true, Public: true, Format: "text", Value: "https://hibiscus.fit"},
		{Key: constants.KEY_SITE_NAME, Desc: "站点名称", Autoload: true, Public: true, Format: "text", Value: "VoiceSculptor"},
		{Key: constants.KEY_SITE_LOGO_URL, Desc: "站点Logo", Autoload: true, Public: true, Format: "text", Value: "/static/img/favicon.png"},
		{Key: constants.KEY_SITE_DESCRIPTION, Desc: "站点描述", Autoload: true, Public: true, Format: "text", Value: "VoiceSculptor - 智能语音客服平台"},
		{Key: constants.KEY_SITE_SIGNIN_URL, Desc: "登录页面", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/login"},
		{Key: constants.KEY_SITE_FAVICON_URL, Desc: "站点图标", Autoload: true, Public: true, Format: "text", Value: "/static/img/favicon.png"},
		{Key: constants.KEY_SITE_SIGNUP_URL, Desc: "注册页面", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/register"},
		{Key: constants.KEY_SITE_LOGOUT_URL, Desc: "注销页面", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/logout"},
		{Key: constants.KEY_SITE_RESET_PASSWORD_URL, Desc: "重置密码页面", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/reset-password"},
		{Key: constants.KEY_SITE_SIGNIN_API, Desc: "登录接口", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/login"},
		{Key: constants.KEY_SITE_SIGNUP_API, Desc: "注册接口", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/register"},
		{Key: constants.KEY_SITE_RESET_PASSWORD_DONE_API, Desc: "重置密码接口", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/auth/reset-password-done"},
		{Key: constants.KEY_SITE_LOGIN_NEXT, Desc: "登录成功后跳转页面", Autoload: true, Public: true, Format: "text", Value: config.GlobalConfig.APIPrefix + "/admin"},
		{Key: constants.KEY_SITE_USER_ID_TYPE, Desc: "用户ID类型", Autoload: true, Public: true, Format: "text", Value: "email"},
		{Key: constants.KEY_SITE_TERMS_URL, Desc: "服务条款", Autoload: true, Public: true, Format: "text", Value: "https://hibiscus.fit"},
	}
	for _, cfg := range defaults {
		var count int64
		err := db.Model(&util.Config{}).Where("`key` = ?", cfg.Key).Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&cfg).Error; err != nil {
				return err
			}
		}
	}

	defaultAdmin := []models.User{
		{
			Email:       "admin@hibiscus.fit",
			Password:    models.HashPassword("admin123"),
			IsStaff:     true,
			IsSuperUser: true,
			DisplayName: "管理员",
			Enabled:     true,
		},
	}
	for _, user := range defaultAdmin {
		var count int64
		err := db.Model(&models.User{}).Where("`email` = ?", user.Email).Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&user).Error; err != nil {
				return err
			}
		}
	}

	defaultPrompts := []models.PromptModel{
		{
			Name:        "summarize_article",
			Description: "总结文章的主要内容，适合长文段或博客文章提炼摘要。",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "translate_text",
			Description: "将输入文本翻译为指定语言，适合中英文互译等场景。",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "generate_title",
			Description: "根据文章内容生成简洁有吸引力的标题。",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "email_reply_generator",
			Description: "根据邮件内容和意图自动生成专业的邮件回复。",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	db.Model(models.PromptModel{}).Create(defaultPrompts)

	defaultArgs := []models.PromptArgModel{
		// summarize_article
		{Name: "content", Description: "待总结的文章内容", Required: true, PromptID: 1},

		// translate_text
		{Name: "text", Description: "要翻译的文本", Required: true, PromptID: 2},
		{Name: "target_language", Description: "目标语言（如 en、zh）", Required: true, PromptID: 2},

		// generate_title
		{Name: "article", Description: "文章内容", Required: true, PromptID: 3},

		// email_reply_generator
		{Name: "email_body", Description: "原始邮件内容", Required: true, PromptID: 4},
		{Name: "tone", Description: "回复语气（如正式、轻松）", Required: false, PromptID: 4},
	}
	db.Model(models.PromptArgModel{}).Create(defaultArgs)
	return nil
}

func (app *VoiceSculptorApp) RegisterRoutes(r *gin.Engine) {
	app.handlers.Register(r)
}

func main() {
	if err := printBannerFromFile("banner.txt"); err != nil {
		log.Fatalf("unload banner: %v", err)
	}

	// 1. parse command line parameters
	mode := flag.String("mode", "test", "running environment (development, test, production)")
	flag.Parse()

	// 2. set environment variables
	if *mode != "" {
		os.Setenv("APP_ENV", *mode)
	}

	// 3. load global configuration
	if err := config.Load(); err != nil {
		panic("config load failed: " + err.Error())
	}

	// 4. load log configuration
	err := logger.Init(&config.GlobalConfig.Log, config.GlobalConfig.Mode)
	if err != nil {
		panic(err)
	}

	// 5. record configuration information
	logger.Info("system config load finished",

		// base
		zap.Int64("machine_id", config.GlobalConfig.MachineID),
		zap.String("addr", config.GlobalConfig.Addr),
		zap.String("db_driver", config.GlobalConfig.DBDriver),
		zap.String("dsn", config.GlobalConfig.DSN),
		zap.String("mode", config.GlobalConfig.Mode),

		zap.String("api_prefix", config.GlobalConfig.APIPrefix),
		zap.String("docs_prefix", config.GlobalConfig.DocsPrefix),
		zap.String("admin_prefix", config.GlobalConfig.AdminPrefix),
		zap.String("auth_prefix", config.GlobalConfig.AuthPrefix),
		zap.String("secret_expire_days", config.GlobalConfig.SecretExpireDays),
		zap.String("session_secret", config.GlobalConfig.SessionSecret),

		// logger
		zap.String("log_level", config.GlobalConfig.Log.Level),
		zap.String("log_filename", config.GlobalConfig.Log.Filename),
		zap.Int("log_max_size", config.GlobalConfig.Log.MaxSize),
		zap.Int("log_max_age", config.GlobalConfig.Log.MaxAge),
		zap.Int("log_max_backups", config.GlobalConfig.Log.MaxBackups),

		// mail
		zap.String("mail_host", config.GlobalConfig.Mail.Host),
		zap.String("mail_username", config.GlobalConfig.Mail.Username),
		zap.String("mail_password", config.GlobalConfig.Mail.Password),
		zap.String("mail_from", config.GlobalConfig.Mail.From),
		zap.Int64("mail_port", config.GlobalConfig.Mail.Port),

		// RustPBX
		zap.String("rust_pbx_url", config.GlobalConfig.RustPbxUrl),
		zap.String("rust_pbx_websocket_url", config.GlobalConfig.RustPbxWebSocketURL),
	)

	// 6. load data source
	logWriter := os.Stdout
	dbDriver := config.GlobalConfig.DBDriver
	dsn := config.GlobalConfig.DSN
	db, err := util.InitDatabase(logWriter, dbDriver, dsn)
	if err != nil {
		logger.Error("init database failed: ", zap.Error(err))
	}

	// 7. load models
	err = util.MakeMigrates(db, []any{
		&util.Config{},
		&models.User{},
		&models.Group{},
		&models.UserCredential{},
		&models.GroupMember{},
		&models.Assistant{},
		&models.ChatSessionLog{},
		&models.PromptModel{},
		&models.PromptArgModel{},
		&notification.InternalNotification{},
	})
	if err != nil {
		logger.Error("migration failed: ", zap.Error(err))
	} else {
		logger.Info("migration success", zap.String("database", dbDriver), zap.String("dsn", dsn))
	}

	if os.Getenv("APP_ENV") != "production" {
		if err := initDefaultConfigs(db); err != nil {
			logger.Error("init default config failed: ", zap.Error(err))
		}
	}

	// 8. load base configs
	var addr = config.GlobalConfig.Addr
	if addr == "" {
		addr = ":8000"
	}

	var DBDriver = config.GlobalConfig.DBDriver
	if DBDriver == "" {
		DBDriver = "sqlite"
	}

	var DSN = config.GlobalConfig.DSN
	if DSN == "" {
		DSN = "file::memory:?cache=shared"
	}
	flag.StringVar(&addr, "addr", addr, "HTTP Serve address")
	flag.StringVar(&DBDriver, "db-driver", DBDriver, "database driver")
	flag.StringVar(&DSN, "dsn", DSN, "database source name")

	logger.Info("checked config -- addr: ", zap.String("addr", addr))
	logger.Info("checked config -- db-driver: ", zap.String("db-driver", DBDriver), zap.String("dsn", DSN))
	logger.Info("checked config -- mode: ", zap.String("mode", config.GlobalConfig.Mode))

	// 9. Load Global Cache
	util.InitGlobalCache(1024, 5*time.Minute)

	//10. Load
	err = prompt.InitPromptSystem(db)
	if err != nil {
		logger.Error("init prompt system failed: ", zap.Error(err))
	}

	// 10. New App
	app := NewVoiceSculptorApp(db)

	// 11. Start timed task
	go task.StartOfflineChecker(db)

	// 12. Initialize gin routing
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("templates/**/**")

	// 13. use middleware

	// Cookie Register
	secret := util.GetEnv(constants.ENV_SESSION_SECRET)
	if secret != "" {
		expireDays := util.GetIntEnv(constants.ENV_SESSION_EXPIRE_DAYS)
		if expireDays <= 0 {
			expireDays = 7
		}
		r.Use(middleware.WithCookieSession(secret, int(expireDays)*24*3600))
	} else {
		r.Use(middleware.WithMemSession(util.RandText(32)))
	}

	// Cors Handle Middleware
	r.Use(middleware.CorsMiddleware())

	// Logger Handle Middleware
	r.Use(middleware.LoggerMiddleware(zap.L()))

	// RateLimit Middleware
	r.Use(middleware.RateLimiterMiddleware())

	// Assets Middleware
	r.Use(voiceSculptor.WithStaticAssets(r, util.GetEnv(constants.ENV_STATIC_PREFIX), util.GetEnv(constants.ENV_STATIC_ROOT)))

	// 14. Register Routes
	app.RegisterRoutes(r)

	// 15. Initialize User Listener
	listeners.InitUserListeners()

	logger.Info("server run success", zap.String("addr", addr))
	// 16. Start HTTP Server
	if err := r.Run(addr); err != nil {
		logger.Error("server run failed", zap.Error(err))
	}
}

func printBannerFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")

	colors := []string{
		"\x1b[38;5;117m",
		"\x1b[38;5;141m",
		"\x1b[38;5;165m",
		"\x1b[38;5;189m",
		"\x1b[38;5;207m",
		"\x1b[38;5;219m",
		"\x1b[38;5;225m",
		"\x1b[38;5;231m",
	}

	for i, line := range lines {
		color := colors[i%len(colors)]
		fmt.Println(color + line + "\x1b[0m")
	}
	return nil
}
