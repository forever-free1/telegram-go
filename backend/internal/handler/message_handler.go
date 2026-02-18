package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/telegram-go/backend/internal/dto"
	"github.com/telegram-go/backend/internal/service"
	"github.com/telegram-go/backend/internal/websocket"
)

type MessageHandler struct {
	messageService *service.MessageService
	hub           *websocket.Hub
}

func NewMessageHandler(messageService *service.MessageService, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{messageService: messageService, hub: hub}
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

// AckMessageRequest 消息已读确认请求
type AckMessageRequest struct {
	MessageIDs []int64 `json:"message_ids" binding:"required,min=1"`
	ChatID    int64   `json:"chat_id" binding:"required"`
}

// AckMessage 处理消息已读确认
// POST /api/messages/ack
func (h *MessageHandler) AckMessage(c *gin.Context) {
	var req AckMessageRequest
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

	// 调用 Service 处理已读确认
	readMessages, err := h.messageService.AckMessages(c.Request.Context(), currentUser.UserID, req.ChatID, req.MessageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	// 通过 WebSocket 通知消息发送者
	if h.hub != nil && len(readMessages) > 0 {
		for _, msg := range readMessages {
			// 通知消息发送者其消息已被读
			readReceipt := &websocket.WSMessage{
				Type:      "WS_MSG_READ",
				MessageID: msg.ID,
				ChatID:    msg.ChatID,
				SenderID:  currentUser.UserID, // 已读者
			}
			if msg.ReadAt != nil {
				readReceipt.Timestamp = *msg.ReadAt
			}
			h.hub.SendToUser(msg.SenderID, readReceipt)
		}
	}

	c.JSON(http.StatusOK, dto.Success(map[string]int{
		"acknowledged": len(readMessages),
	}))
}
