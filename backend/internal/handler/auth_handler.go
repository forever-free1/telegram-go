package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forever-free1/telegram-go/backend/internal/dto"
	"github.com/forever-free1/telegram-go/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 200 {object} dto.Response
// @Router /api/auth/register [post]
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

// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.Response
// @Router /api/auth/login [post]
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

// @Summary User logout
// @Description Logout current user and invalidate token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		h.authService.Logout(c.Request.Context(), token)
	}
	c.JSON(http.StatusOK, dto.Success(nil))
}

// @Summary Get current user
// @Description Get current authenticated user information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /api/user/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	claims, ok := userVal.(*service.UserClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	// Get full user data from database
	user, err := h.authService.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(user))
}
