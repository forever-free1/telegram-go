package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/telegram-go/backend/internal/dto"
	"github.com/telegram-go/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), &service.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
		Email:    req.Email,
		Nickname: req.Nickname,
	})
	if err != nil {
		code := 500
		message := err.Error()
		if err == service.ErrUserAlreadyExists {
			code = 409
			message = "Username already exists"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		code := 500
		message := err.Error()
		if err == service.ErrUserNotFound || err == service.ErrInvalidPassword {
			code = 401
			message = "Invalid username or password"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		h.authService.Logout(c.Request.Context(), token)
	}
	c.JSON(http.StatusOK, dto.Success(nil))
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}
	c.JSON(http.StatusOK, dto.Success(user))
}
