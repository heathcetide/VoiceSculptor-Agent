package handlers

import (
	"VoiceSculptor/internal/models"
	"VoiceSculptor/pkg/response"
	"github.com/gin-gonic/gin"
)

// UpdateRateLimiterConfig 更新限流配置
func (h *Handlers) handleCreateCredential(c *gin.Context) {
	var credential models.UserCredentialRequest
	if err := c.ShouldBindJSON(&credential); err != nil {
		response.Fail(c, "Invalid request", nil)
		return
	}

	user := models.CurrentUser(c)
	if user == nil {
		response.Fail(c, "User is not logged in.", nil)
	}

	userCredential, err := models.CreateUserCredential(h.db, user.ID, &credential)
	if err != nil {
		response.Fail(c, "create user credential failed", err)
		return
	}

	response.Success(c, "create user credential success", gin.H{
		"apiKey":    userCredential.APIKey,
		"apiSecret": userCredential.APISecret,
		"name":      credential.Name,
	})
}

func (h *Handlers) handleGetCredential(c *gin.Context) {
	user := models.CurrentUser(c)
	credentials, err := models.GetUserCredentials(h.db, user.ID)
	if err != nil {
		response.Fail(c, "get user credentials failed", err)
		return
	}
	response.Success(c, "get user credentials success", credentials)
}
