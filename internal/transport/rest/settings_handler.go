package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/shvdev1/HackNeChange/api-gateway/internal/entity"
	redisstore "github.com/shvdev1/HackNeChange/api-gateway/internal/storage/redis"
)

// SettingsHandler holds dependencies for settings transport
type SettingsHandler struct {
	store *redisstore.SettingsStorage
}

// NewSettingsHandler creates a new handler instance
func NewSettingsHandler(store *redisstore.SettingsStorage) *SettingsHandler {
	return &SettingsHandler{store: store}
}

// saveRequest represents the expected JSON body for POST /settings
type saveRequest struct {
	UserID      string `json:"user_id" binding:"omitempty"`
	Theme       string `json:"theme" binding:"required_with=UserID"`
	PickedModel string `json:"picked_model" binding:"required_with=UserID"`
	Font        string `json:"font" binding:"required_with=UserID"`
}

// RegisterRoutes registers the settings routes on the provided router group
func (h *SettingsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/settings", h.SaveSettings)
	rg.GET("/settings/:user_id", h.GetSettings)
}

// SaveSettings handles POST /api/v1/settings
func (h *SettingsHandler) SaveSettings(c *gin.Context) {
	var req saveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Allow user id in header if not provided in body
	userID := req.UserID
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required (body or X-User-ID header)"})
		return
	}

	settings := entity.UserSettings{
		Theme:       req.Theme,
		PickedModel: req.PickedModel,
		Font:        req.Font,
	}

	if err := h.store.SaveSettings(c.Request.Context(), userID, settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user_id": userID, "settings": settings})
}

// GetSettings handles GET /api/v1/settings/:user_id
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	settings, err := h.store.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get settings", "details": err.Error()})
		return
	}

	// Return defaults if no settings exist (GetSettings returns empty struct)
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "settings": settings})
}
