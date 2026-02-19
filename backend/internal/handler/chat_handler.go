package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forever-free1/telegram-go/backend/internal/dto"
	"github.com/forever-free1/telegram-go/backend/internal/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// @Summary Create a chat
// @Description Create a new chat (private or group)
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateChatRequest true "Create chat request"
// @Success 200 {object} dto.Response
// @Router /api/chats [post]
func (h *ChatHandler) CreateChat(c *gin.Context) {
	var req dto.CreateChatRequest
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
	chat, err := h.chatService.CreateChat(c.Request.Context(), currentUser.UserID, &service.CreateChatRequest{
		Name: req.Name,
		Type: req.Type,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(chat))
}

// @Summary Get a chat
// @Description Get chat details by ID
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Success 200 {object} dto.Response
// @Router /api/chats/{id} [get]
func (h *ChatHandler) GetChat(c *gin.Context) {
	var req struct {
		ChatID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	chat, err := h.chatService.GetChat(c.Request.Context(), req.ChatID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Error(404, "chat not found"))
		return
	}

	c.JSON(http.StatusOK, dto.Success(chat))
}

// @Summary Get user chats
// @Description Get all chats for current user
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /api/chats [get]
func (h *ChatHandler) GetUserChats(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)
	chats, err := h.chatService.GetUserChats(c.Request.Context(), currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(chats))
}

// @Summary Add member to chat
// @Description Add a user to a chat
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddChatMemberRequest true "Add member request"
// @Success 200 {object} dto.Response
// @Router /api/chats/members [post]
func (h *ChatHandler) AddMember(c *gin.Context) {
	var req dto.AddChatMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	err := h.chatService.AddMember(c.Request.Context(), req.ChatID, req.UserID, 1)
	if err != nil {
		code := 500
		message := err.Error()
		if err == service.ErrUserNotFound {
			code = 404
			message = "User not found"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// @Summary Remove member from chat
// @Description Remove a user from a chat
// @Tags chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]int64 true "Remove member request"
// @Success 200 {object} dto.Response
// @Router /api/chats/members [delete]
func (h *ChatHandler) RemoveMember(c *gin.Context) {
	var req struct {
		ChatID int64 `json:"chat_id" binding:"required"`
		UserID int64 `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	err := h.chatService.RemoveMember(c.Request.Context(), req.ChatID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// @Summary Get chat members
// @Description Get all members of a chat
// @Tags chats
// @Produce json
// @Security BearerAuth
// @Param id path int true "Chat ID"
// @Success 200 {object} dto.Response
// @Router /api/chats/{id}/members [get]
func (h *ChatHandler) GetMembers(c *gin.Context) {
	var req struct {
		ChatID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	members, err := h.chatService.GetMembers(c.Request.Context(), req.ChatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(members))
}
