package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/telegram-go/backend/internal/dto"
	"github.com/telegram-go/backend/internal/service"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
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
	message, err := h.messageService.SendMessage(c.Request.Context(), currentUser.UserID, &service.SendMessageRequest{
		ChatID:    req.ChatID,
		Type:      req.Type,
		Content:   req.Content,
		MediaURL:  req.MediaURL,
		Duration:  req.Duration,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		ReplyID:   req.ReplyID,
	})
	if err != nil {
		code := 500
		message := err.Error()
		switch err {
		case service.ErrChatNotFound:
			code = 404
			message = "Chat not found"
		case service.ErrNotAuthorized:
			code = 403
			message = "Not authorized"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(message))
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	var req dto.GetMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)
	messages, err := h.messageService.GetMessages(c.Request.Context(), req.ChatID, currentUser.UserID, req.Offset, req.Limit)
	if err != nil {
		code := 500
		message := err.Error()
		switch err {
		case service.ErrChatNotFound:
			code = 404
			message = "Chat not found"
		case service.ErrNotAuthorized:
			code = 403
			message = "Not authorized"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(messages))
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	var req struct {
		MessageID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)
	err := h.messageService.DeleteMessage(c.Request.Context(), req.MessageID, currentUser.UserID)
	if err != nil {
		code := 500
		message := err.Error()
		switch err {
		case service.ErrMessageNotFound:
			code = 404
			message = "Message not found"
		case service.ErrNotAuthorized:
			code = 403
			message = "Not authorized"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}
