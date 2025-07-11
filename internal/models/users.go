package models

import (
	voiceSculptor "VoiceSculptor"
	constants "VoiceSculptor/pkg/constant"
	"VoiceSculptor/pkg/util"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	//SigUserLogin: user *User, c *gin.Context
	SigUserLogin = "user.login"
	//SigUserLogout: user *User, c *gin.Context
	SigUserLogout = "user.logout"
	//SigUserCreate: user *User, c *gin.Context
	SigUserCreate = "user.create"
	//SigUserVerifyEmail: user *User, hash, clientIp, userAgent string
	SigUserVerifyEmail = "user.verifyemail"
	//SigUserResetPassword: user *User, hash, clientIp, userAgent string
	SigUserResetPassword = "user.resetpassword"
)

type SendEmailVerifyEmail struct {
	Email     string `json:"email"`
	ClientIp  string `json:"clientIp"`
	UserAgent string `json:"userAgent"`
}

type LoginForm struct {
	Email     string `json:"email" comment:"Email address"`
	Password  string `json:"password,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
	Remember  bool   `json:"remember,omitempty"`
	AuthToken string `json:"token,omitempty"`
}

type EmailOperatorForm struct {
	Email     string `json:"email" comment:"Email address"`
	Code      string `json:"code"`
	AuthToken bool   `json:"AuthToken,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
}

type RegisterUserForm struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"displayName"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Locale      string `json:"locale"`
	Timezone    string `json:"timezone"`
	Source      string `json:"source"`
}

type ChangePasswordForm struct {
	Password string `json:"password" binding:"required"`
}

type ResetPasswordForm struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordDoneForm struct {
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

type UpdateUserRequest struct {
	Email       string `form:"email" json:"email"`
	Phone       string `form:"phone" json:"phone"`
	DisplayName string `form:"displayName" json:"displayName"`
	Locale      string `form:"locale" json:"locale"`
	Timezone    string `form:"timezone" json:"timezone"`
	Gender      string `form:"gender" json:"gender"`
	Extra       string `form:"extra" json:"extra"`
	Avatar      string `form:"avatar" json:"avatar"`
}

func Login(c *gin.Context, user *User) {
	db := c.MustGet(constants.DbField).(*gorm.DB)
	SetLastLogin(db, user, c.ClientIP())
	session := sessions.Default(c)
	session.Set(constants.UserField, user.ID)
	session.Save()
	util.Sig().Emit(SigUserLogin, user, c)
}

func Logout(c *gin.Context, user *User) {
	c.Set(constants.UserField, nil)
	session := sessions.Default(c)
	session.Delete(constants.UserField)
	session.Save()
	util.Sig().Emit(SigUserLogout, user, c)
}

func AuthRequired(c *gin.Context) {
	if CurrentUser(c) != nil {
		c.Next()
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, errors.New("authorization required"))
		return
	}

	db := c.MustGet(constants.DbField).(*gorm.DB)
	// split bearer
	token = strings.TrimPrefix(token, "Bearer ")
	user, err := DecodeHashToken(db, token, false)
	if err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, err)
		return
	}
	c.Set(constants.UserField, user)
	c.Next()
}

func AuthApiRequired(c *gin.Context) {
	if CurrentUser(c) != nil {
		c.Next()
		return
	}

	apiKey := c.GetHeader("X-API-KEY")
	apiSecret := c.GetHeader("X-API-SECRET")
	if apiKey != "" && apiSecret != "" {
		user, err := GetUserByAPIKey(c, apiKey, apiSecret)
		if err != nil {
			voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, err)
			return
		}
		c.Set(constants.UserField, user)
		c.Next()
		return
	}

	apiKey = c.Query("apiKey")
	apiSecret = c.Query("apiSecret")
	if apiKey != "" && apiSecret != "" {
		user, err := GetUserByAPIKey(c, apiKey, apiSecret)
		if err != nil {
			voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, err)
			return
		}
		c.Set(constants.UserField, user)
		c.Next()
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, errors.New("authorization required"))
		return
	}

	db := c.MustGet(constants.DbField).(*gorm.DB)
	// split bearer
	token = strings.TrimPrefix(token, "Bearer ")
	user, err := DecodeHashToken(db, token, false)
	if err != nil {
		voiceSculptor.AbortWithJSONError(c, http.StatusUnauthorized, err)
		return
	}
	c.Set(constants.UserField, user)
	c.Next()
}

func GetUserByAPIKey(c *gin.Context, apiKey, apiSecret string) (*User, error) {
	db := c.MustGet(constants.DbField).(*gorm.DB)
	var userCredential UserCredential
	err := db.Model(&UserCredential{}).Where("api_key = ? AND api_secret = ?", apiKey, apiSecret).Find(&userCredential).Error
	if err != nil {
		return nil, err
	}
	var user *User
	err = db.Model(&User{}).Where("id = ?", userCredential.UserID).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CurrentUser(c *gin.Context) *User {
	if cachedObj, exists := c.Get(constants.UserField); exists && cachedObj != nil {
		return cachedObj.(*User)
	}
	session := sessions.Default(c)
	userId := session.Get(constants.UserField)
	if userId == nil {
		return nil
	}

	db := c.MustGet(constants.DbField).(*gorm.DB)
	user, err := GetUserByUID(db, userId.(uint))
	if err != nil {
		return nil
	}
	c.Set(constants.UserField, user)
	return user
}

func CheckPassword(user *User, password string) bool {
	if user.Password == "" {
		return false
	}
	return user.Password == HashPassword(password)
}

func SetPassword(db *gorm.DB, user *User, password string) (err error) {
	p := HashPassword(password)
	err = UpdateUserFields(db, user, map[string]any{
		"Password": p,
	})
	if err != nil {
		return
	}
	user.Password = p
	return
}

func HashPassword(password string) string {
	if password == "" {
		return ""
	}
	hashVal := sha256.Sum256([]byte(password))
	return fmt.Sprintf("sha256$%x", hashVal)
}

func GetUserByUID(db *gorm.DB, userID uint) (*User, error) {
	var val User
	result := db.Where("id", userID).Where("enabled", true).Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func GetUserByEmail(db *gorm.DB, email string) (user *User, err error) {
	var val User
	result := db.Table("users").Where("email", strings.ToLower(email)).Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func IsExistsByEmail(db *gorm.DB, email string) bool {
	_, err := GetUserByEmail(db, email)
	return err == nil
}

func CreateUser(db *gorm.DB, email, password string) (*User, error) {
	user := User{
		Email:     email,
		Password:  HashPassword(password),
		Enabled:   true,
		Activated: false,
	}

	result := db.Create(&user)
	return &user, result.Error
}
func UpdateUserFields(db *gorm.DB, user *User, vals map[string]any) error {
	return db.Model(user).Updates(vals).Error
}

func SetLastLogin(db *gorm.DB, user *User, lastIp string) error {
	now := time.Now().Truncate(1 * time.Second)
	vals := map[string]any{
		"LastLoginIP": lastIp,
		"LastLogin":   &now,
	}
	user.LastLogin = &now
	user.LastLoginIP = lastIp
	return db.Model(user).Updates(vals).Error
}

func EncodeHashToken(user *User, timestamp int64, useLastlogin bool) (hash string) {
	//
	// ts-uid-token
	logintimestamp := "0"
	if useLastlogin && user.LastLogin != nil {
		logintimestamp = fmt.Sprintf("%d", user.LastLogin.Unix())
	}
	t := fmt.Sprintf("%s$%d", user.Email, timestamp)
	hashVal := sha256.Sum256([]byte(logintimestamp + user.Password + t))
	hash = base64.RawStdEncoding.EncodeToString([]byte(t)) + "-" + fmt.Sprintf("%x", hashVal)
	return hash
}

func DecodeHashToken(db *gorm.DB, hash string, useLastLogin bool) (user *User, err error) {
	vals := strings.Split(hash, "-")
	if len(vals) != 2 {
		return nil, errors.New("bad token")
	}
	data, err := base64.RawStdEncoding.DecodeString(vals[0])
	if err != nil {
		return nil, errors.New("bad token")
	}

	vals = strings.Split(string(data), "$")
	if len(vals) != 2 {
		return nil, errors.New("bad token")
	}

	ts, err := strconv.ParseInt(vals[1], 10, 64)
	if err != nil {
		return nil, errors.New("bad token")
	}

	if time.Now().Unix() > ts {
		return nil, errors.New("token expired")
	}

	user, err = GetUserByEmail(db, vals[0])
	if err != nil {
		return nil, errors.New("bad token")
	}
	token := EncodeHashToken(user, ts, useLastLogin)
	if token != hash {
		return nil, errors.New("bad token")
	}
	return user, nil
}

func CheckUserAllowLogin(db *gorm.DB, user *User) error {
	if !user.Enabled {
		return errors.New("user not allow login")
	}

	if util.GetBoolValue(db, constants.KEY_USER_ACTIVATED) && !user.Activated {
		return errors.New("waiting for activation")
	}
	return nil
}

func InTimezone(c *gin.Context, timezone string) {
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return
	}
	c.Set(constants.TzField, tz)

	session := sessions.Default(c)
	session.Set(constants.TzField, timezone)
	session.Save()
}

func BuildAuthToken(user *User, expired time.Duration, useLoginTime bool) string {
	n := time.Now().Add(expired)
	return EncodeHashToken(user, n.Unix(), useLoginTime)
}

func UpdateUser(db *gorm.DB, user *User, vals map[string]any) error {
	return db.Model(user).Updates(vals).Error
}
