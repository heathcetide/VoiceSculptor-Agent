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
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	db := c.MustGet(constants.DbField).(*gorm.DB)
	var err error
	user, err = models.GetUserByUID(db, user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	withToken := c.Query("with_token")
	if withToken != "" {
		expired, err := time.ParseDuration(withToken)
		if err == nil {
			if expired >= 24*time.Hour {
				expired = 24 * time.Hour
			}
			user.AuthToken = models.BuildAuthToken(db, user, expired, false)
		}
	}
	c.JSON(http.StatusOK, user)
}

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
		user.AuthToken = models.BuildAuthToken(db, user, expired, false)
	}
	c.JSON(http.StatusOK, user)
}

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

	util.Sig().Emit(models.SigUserCreate, user, c)

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

// Send Email Code
func (h *Handlers) handleSendEmailCode(context *gin.Context) {
	var req models.SendEmailVerifyEmail
	req.UserAgent = context.Request.UserAgent()
	req.ClientIp = context.ClientIP()
	if err := context.BindJSON(&req); err != nil {
		voiceSculptor.AbortWithJSONError(context, http.StatusBadRequest, err)
		return
	}
	text := util.RandNumberText(6)
	util.GlobalCache.Add(req.Email, text)
	err := notification.NewMailNotification(config.GlobalConfig.Mail).SendVerificationCode(req.Email, text)
	if err != nil {
		voiceSculptor.AbortWithJSONError(context, http.StatusBadRequest, err)
		return
	}
	response.Success(context, "success", "Send Email Successful, Must be verified within the valid time [5 minutes]")
}
