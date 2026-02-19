package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forever-free1/telegram-go/backend/internal/dto"
	"github.com/forever-free1/telegram-go/backend/internal/service"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	notificationService *service.NotificationService
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler(notificationService *service.NotificationService) *DeviceHandler {
	return &DeviceHandler{notificationService: notificationService}
}

// RegisterTokenRequest 注册 Token 请求
type RegisterTokenRequest struct {
	FCMToken   string `json:"fcm_token" binding:"required"`
	DeviceType string `json:"device_type" binding:"required"` // ios, android, web
}

// RegisterToken 注册设备 Token
// POST /api/device/token
func (h *DeviceHandler) RegisterToken(c *gin.Context) {
	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	err := h.notificationService.RegisterToken(
		c.Request.Context(),
		currentUser.UserID,
		req.FCMToken,
		req.DeviceType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, "failed to register token"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(map[string]string{
		"status": "registered",
	}))
}

// UnregisterToken 注销设备 Token
// DELETE /api/device/token
func (h *DeviceHandler) UnregisterToken(c *gin.Context) {
	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	err := h.notificationService.UnregisterToken(
		c.Request.Context(),
		currentUser.UserID,
		req.FCMToken,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, "failed to unregister token"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(map[string]string{
		"status": "unregistered",
	}))
}
