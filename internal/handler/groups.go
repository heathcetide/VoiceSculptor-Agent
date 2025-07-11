package handlers

import (
	"VoiceSculptor/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreateGroupRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Type       string                 `json:"type"`
	Extra      string                 `json:"extra"`
	Permission models.GroupPermission `json:"permission"`
}

func (h *Handlers) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := models.Group{
		Name:       req.Name,
		Type:       req.Type,
		Extra:      req.Extra,
		Permission: req.Permission,
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *Handlers) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group

	if err := h.db.Delete(&group, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func (h *Handlers) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var req CreateGroupRequest
	var group models.Group

	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.First(&group, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	group.Name = req.Name
	group.Type = req.Type
	group.Extra = req.Extra
	group.Permission = req.Permission

	if err := h.db.Save(&group).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *Handlers) GetGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group

	if err := h.db.First(&group, id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *Handlers) ListGroups(c *gin.Context) {
	var groups []models.Group
	if err := h.db.Find(&groups).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}
