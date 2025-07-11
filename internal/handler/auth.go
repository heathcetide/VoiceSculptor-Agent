package handlers

import (
	voiceSculptor "VoiceSculptor"
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/config"
	constants "VoiceSculptor/pkg/constant"
	"VoiceSculptor/pkg/logger"
	"VoiceSculptor/pkg/notification"
	"VoiceSculptor/pkg/response"
	"VoiceSculptor/pkg/util"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func (h *Handlers) handleUserSignupPage(c *gin.Context) {
	ctx := voiceSculptor.GetRenderPageContext(c)
	ctx["SignupText"] = "Sign Up Now"
	ctx["Site.SignupApi"] = util.GetValue(h.db, constants.KEY_SITE_SIGNUP_API)
	c.HTML(http.StatusOK, "signup.html", ctx)
}

func (h *Handlers) handleUserResetPasswordPage(c *gin.Context) {
	c.HTML(http.StatusOK, "reset_password.html", voiceSculptor.GetRenderPageContext(c))
}

func (h *Handlers) handleUserSigninPage(c *gin.Context) {
	ctx := voiceSculptor.GetRenderPageContext(c)
	ctx["SignupText"] = "Sign Up Now"
	c.HTML(http.StatusOK, "signin.html", ctx)
}

func (h *Handlers) handleUserLogout(c *gin.Context) {
	user := models.CurrentUser(c)
	if user != nil {
		models.Logout(c, user)
	}
	next := c.Query("next")
	if next != "" {
		c.Redirect(http.StatusFound, next)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handlers) handleUserInfo(c *gin.Context) {
	user := models.CurrentUser(c)
	if user == nil {
		response.AbortWithStatus(c, http.StatusUnauthorized)
		return
	}
	db := c.MustGet(constants.DbField).(*gorm.DB)
	var err error
	user, err = models.GetUserByUID(db, user.ID)
	if err != nil {
		response.AbortWithStatus(c, http.StatusUnauthorized)
		return
	}
	withToken := c.Query("with_token")
	if withToken != "" {
		expired, err := time.ParseDuration(withToken)
		if err == nil {
			if expired >= 24*time.Hour {
				expired = 24 * time.Hour
			}
			user.AuthToken = models.BuildAuthToken(user, expired, false)
		}
	}
	response.Success(c, "success", user)
}

func (h *Handlers) handleUserSigninByEmail(c *gin.Context) {
	var form models.EmailOperatorForm
	if err := c.BindJSON(&form); err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	// 检查邮箱是否为空
	if form.Email == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("email is required"))
		return
	}

	// 获取数据库实例
	db := c.MustGet(constants.DbField).(*gorm.DB)

	// 获取用户
	user, err := models.GetUserByEmail(db, form.Email)
	if err != nil {
		response.Fail(c, "user not exists", errors.New("user not exists"))
		return
	}

	// 校验验证码
	if form.Code == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("verification code is required"))
		return
	}

	// 从缓存中获取验证码（假设你使用的是 util.GlobalCache）
	cachedCode, ok := util.GlobalCache.Get(form.Email)
	if !ok || cachedCode != form.Code {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("invalid verification code"))
		return
	}

	// 清除已用验证码
	util.GlobalCache.Remove(form.Email)

	// 检查用户是否允许登录（激活、启用等）
	err = models.CheckUserAllowLogin(db, user)
	if err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusForbidden, err)
		return
	}

	// 设置时区（如果有的话）
	if form.Timezone != "" {
		models.InTimezone(c, form.Timezone)
	}

	// 登录用户，设置 Session
	models.Login(c, user)

	// 如果需要 Token，生成 AuthToken
	if form.AuthToken {
		val := util.GetValue(db, constants.KEY_AUTH_TOKEN_EXPIRED)
		expired, _ := time.ParseDuration(val)
		if expired < 24*time.Hour {
			expired = 24 * time.Hour
		}
		user.AuthToken = models.BuildAuthToken(user, expired, false)
	}

	// 返回用户信息
	response.Success(c, "login success", user)
}

// handleUserSignin handle user signin
func (h *Handlers) handleUserSignin(c *gin.Context) {
	var form models.LoginForm
	if err := c.BindJSON(&form); err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	if form.AuthToken == "" && form.Email == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("email is required"))
		return
	}

	if form.Password == "" && form.AuthToken == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("empty password"))
		return
	}

	db := c.MustGet(constants.DbField).(*gorm.DB)
	var user *models.User
	var err error
	if form.Password != "" {
		user, err = models.GetUserByEmail(db, form.Email)
		if err != nil {
			voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("user not exists"))
			return
		}
		if !models.CheckPassword(user, form.Password) {
			voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, errors.New("unauthorized"))
			return
		}
	} else {
		user, err = models.DecodeHashToken(db, form.AuthToken, false)
		if err != nil {
			voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, err)
			return
		}
	}

	err = models.CheckUserAllowLogin(db, user)
	if err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusForbidden, err)
		return
	}

	if form.Timezone != "" {
		models.InTimezone(c, form.Timezone)
	}

	models.Login(c, user)

	if form.Remember {
		val := util.GetValue(db, constants.KEY_AUTH_TOKEN_EXPIRED) // 7d
		expired, err := time.ParseDuration(val)
		if err != nil {
			// 7 days
			expired = 7 * 24 * time.Hour
		}
		user.AuthToken = models.BuildAuthToken(user, expired, false)
	}
	c.JSON(http.StatusOK, user)
}

// handleUserSignup handle user signup
func (h *Handlers) handleUserSignup(c *gin.Context) {
	var form models.RegisterUserForm
	if err := c.BindJSON(&form); err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	db := c.MustGet(constants.DbField).(*gorm.DB)
	if models.IsExistsByEmail(db, form.Email) {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("email has exists"))
		return
	}

	user, err := models.CreateUser(db, form.Email, form.Password)
	if err != nil {
		logger.Warn("create user failed", zap.Any("email", form.Email), zap.Error(err))
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	vals := util.StructAsMap(form, []string{
		"DisplayName",
		"FirstName",
		"LastName",
		"Locale",
		"Timezone",
		"Source"})

	n := time.Now().Truncate(1 * time.Second)
	vals["LastLogin"] = &n
	vals["LastLoginIP"] = c.ClientIP()

	user.DisplayName = form.DisplayName
	user.FirstName = form.FirstName
	user.LastName = form.LastName
	user.Locale = form.Locale
	user.Source = "ADMIN"
	user.Timezone = form.Timezone
	user.LastLogin = &n
	user.LastLoginIP = c.ClientIP()

	err = models.UpdateUserFields(db, user, vals)
	if err != nil {
		logger.Warn("update user fields fail id:", zap.Uint("userId", user.ID), zap.Any("vals", vals), zap.Error(err))
	}

	util.Sig().Emit(models.SigUserCreate, user, c, db)

	r := gin.H{
		"email":      user.Email,
		"activation": user.Activated,
	}
	if !user.Activated && util.GetBoolValue(db, constants.KEY_USER_ACTIVATED) {
		sendHashMail(db, user, models.SigUserVerifyEmail, constants.KEY_VERIFY_EMAIL_EXPIRED, "180d", c.ClientIP(), c.Request.UserAgent())
		r["expired"] = "180d"
	} else {
		models.Login(c, user) //Login now
	}
	c.JSON(http.StatusOK, r)
}

// handleUserSignupByEmail email register email activation
func (h *Handlers) handleUserSignupByEmail(c *gin.Context) {
	var form models.EmailOperatorForm
	if err := c.BindJSON(&form); err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}
	db := c.MustGet(constants.DbField).(*gorm.DB)
	if models.IsExistsByEmail(db, form.Email) {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("email has exists"))
		return
	}
	// 从缓存中获取验证码（假设你使用的是 util.GlobalCache）
	cachedCode, ok := util.GlobalCache.Get(form.Email)
	if !ok || cachedCode != form.Code {
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, errors.New("invalid verification code"))
		return
	}

	// 清除已用验证码
	util.GlobalCache.Remove(form.Email)

	user, err := models.CreateUser(db, form.Email, "123456789")
	if err != nil {
		logger.Warn("create user failed", zap.Any("email", form.Email), zap.Error(err))
		voiceSculptor.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}
	vals := util.StructAsMap(form, []string{
		"DisplayName",
		"FirstName",
		"LastName",
		"Locale",
		"Timezone",
		"Source"})
	user.Source = "ADMIN"
	user.Timezone = form.Timezone
	err = models.UpdateUserFields(db, user, vals)
	if err != nil {
		logger.Warn("update user fields fail id:", zap.Uint("userId", user.ID), zap.Any("vals", vals), zap.Error(err))
	}
	util.Sig().Emit(models.SigUserCreate, user, c)
	go func() {
		err = db.Create(&notification.InternalNotification{
			UserID:    user.ID,
			Title:     "welcome to use the voiceSculptor",
			Content:   "welcome to join VoiceSculptor，now, you can create your first assistant",
			Read:      false,
			CreatedAt: time.Now(),
		}).Error
		if err != nil {
			logger.Warn("send first notification failed  ", zap.Error(err))
		}
		if config.GlobalConfig.DBDriver == "sqlite" {
			db.Model(&notification.InternalNotification{}).Create(notification.InternalNotification{
				UserID:    user.ID,
				Title:     "请注意! 当前您正在使用sqlite作为主数据库",
				Content:   "请将数据库切换到mysql，sqlite仅用于测试, 可能导致数据丢失, 如需上线，建议更换数据库",
				Read:      false,
				CreatedAt: time.Now(),
			})
		}
	}()
	sendHashMail(db, user, models.SigUserVerifyEmail, constants.KEY_VERIFY_EMAIL_EXPIRED, "180d", c.ClientIP(), c.Request.UserAgent())
	response.Success(c, "signup success", user)
}

// handleUserUpdate Update User Info
func (h *Handlers) handleUserUpdate(c *gin.Context) {
	var req models.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, "Invalid request", err)
		return
	}

	user := models.CurrentUser(c)
	vals := make(map[string]interface{})

	if req.Email != "" {
		vals["email"] = req.Email
	}
	if req.Phone != "" {
		vals["phone"] = req.Phone
	}
	if req.DisplayName != "" {
		vals["display_name"] = req.DisplayName
	}
	if req.Locale != "" {
		vals["locale"] = req.Locale
	}
	if req.Timezone != "" {
		vals["timezone"] = req.Timezone
	}
	if req.Gender != "" {
		vals["gender"] = req.Gender
	}
	if req.Extra != "" {
		vals["extra"] = req.Extra
	}
	if req.Avatar != "" {
		vals["avatar"] = req.Avatar
	}

	err := models.UpdateUser(h.db, user, vals)
	if err != nil {
		response.Fail(c, "update user failed", err)
		return
	}
	response.Success(c, "update user success", nil)
}

func (h *Handlers) handleUserUpdatePreferences(c *gin.Context) {
	var preferences struct {
		EmailNotifications bool `json:"emailNotifications"`
	}
	if err := c.ShouldBind(&preferences); err != nil {
		response.Fail(c, "Invalid request", err)
		return
	}
	user := models.CurrentUser(c)
	if err := models.UpdateUser(h.db, user, map[string]any{
		"email_notifications": preferences.EmailNotifications,
	}); err != nil {
		response.Fail(c, "update user failed", err)
		return
	}
	response.Success(c, "Update user preferences successfully", nil)
}

func sendHashMail(db *gorm.DB, user *models.User, signame, expireKey, defaultExpired, clientIp, useragent string) {
	d, err := time.ParseDuration(util.GetValue(db, expireKey))
	if err != nil {
		d, _ = time.ParseDuration(defaultExpired)
	}
	n := time.Now().Add(d)
	hash := models.EncodeHashToken(user, n.Unix(), true)
	// Send Mail
	mailer := notification.NewMailNotification(config.GlobalConfig.Mail)

	err = mailer.SendWelcomeEmail(
		user.Email,
		user.DisplayName,
		"https://yourapp.com/verify?token=abc123", // 验证链接
	)
	if err != nil {
		logger.Warn("send mail failed", zap.Error(err))
		return
	}
	util.Sig().Emit(signame, user, hash, clientIp, useragent)
}

// handleSendEmailCode Send Email Code
func (h *Handlers) handleSendEmailCode(context *gin.Context) {
	var req models.SendEmailVerifyEmail
	if err := context.BindJSON(&req); err != nil {
		voiceSculptor.AbortWithJSONError(context, http.StatusBadRequest, err)
		return
	}
	req.UserAgent = context.Request.UserAgent()
	req.ClientIp = context.ClientIP()
	text := util.RandNumberText(6)
	util.GlobalCache.Add(req.Email, text)
	go func() {
		err := notification.NewMailNotification(config.GlobalConfig.Mail).SendVerificationCode(req.Email, text)
		if err != nil {
			voiceSculptor.AbortWithJSONError(context, http.StatusBadRequest, err)
			return
		}
	}()
	response.Success(context, "success", "Send Email Successful, Must be verified within the valid time [5 minutes]")
	return
}
