package service

import (
	"context"
	"errors"
	"sync"

	"gorm.io/gorm"

	"github.com/telegram-go/backend/internal/model"
	"github.com/telegram-go/backend/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrChatNotFound    = errors.New("chat not found")
	ErrMessageNotFound = errors.New("message not found")
	ErrNotAuthorized  = errors.New("not authorized to perform this action")
)

type UserClaims struct {
	UserID int64
}

// MessageEventHandler 消息事件处理器接口
type MessageEventHandler interface {
	OnMessageSaved(message *model.Message)
}

// MessageBroadcaster 消息广播器，用于在消息保存后触发广播
type MessageBroadcaster interface {
	OnMessageSaved(message *model.Message)
}

var _ MessageBroadcaster = (*MessageBroadcasterFunc)(nil)

// MessageBroadcasterFunc 函数式回调实现
type MessageBroadcasterFunc func(message *model.Message)

func (f MessageBroadcasterFunc) OnMessageSaved(message *model.Message) {
	f(message)
}

type MessageService struct {
	messageRepo  *repository.MessageRepository
	chatRepo     *repository.ChatRepository
	userRepo     *repository.UserRepository
	logger       *zap.Logger
	broadcaster  MessageBroadcaster
	broadcasterMu sync.RWMutex
}

func NewMessageService(
	messageRepo *repository.MessageRepository,
	chatRepo *repository.ChatRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// SetBroadcaster 设置消息广播器
func (s *MessageService) SetBroadcaster(broadcaster MessageBroadcaster) {
	s.broadcasterMu.Lock()
	defer s.broadcasterMu.Unlock()
	s.broadcaster = broadcaster
}

// broadcast 广播消息（线程安全）
func (s *MessageService) broadcast(message *model.Message) {
	s.broadcasterMu.RLock()
	broadcaster := s.broadcaster
	s.broadcasterMu.RUnlock()

	if broadcaster != nil {
		broadcaster.OnMessageSaved(message)
	}
}

type SendMessageRequest struct {
	ChatID    int64   `json:"chat_id" validate:"required"`
	Type      int     `json:"type" validate:"required,min=1,max=5"`
	Content   string  `json:"content"`
	MediaURL  string  `json:"media_url"`
	Duration  int     `json:"duration"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	ReplyID   int64   `json:"reply_id"`
}

func (s *MessageService) SendMessage(ctx context.Context, senderID int64, req *SendMessageRequest) (*model.Message, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindByID(ctx, req.ChatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	// Check if user is a member of the chat
	_, err = s.chatRepo.GetMember(ctx, req.ChatID, senderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotAuthorized
		}
		return nil, err
	}

	// Generate SeqID (using snowflake or auto-increment)
	// Here we use 0 and let database auto-increment
	// In production, you might want to use a distributed ID generator

	// Create message
	message := &model.Message{
		SeqID:     0, // Will be auto-generated
		ChatID:    req.ChatID,
		SenderID:  senderID,
		Type:      req.Type,
		Content:   req.Content,
		MediaURL:  req.MediaURL,
		Duration:  req.Duration,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		ReplyID:   req.ReplyID,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		s.logger.Error("failed to create message", zap.Error(err))
		return nil, err
	}

	// 消息保存成功后，触发广播
	s.broadcast(message)

	return message, nil
}

// SendMessageFromWS 从 WebSocket 发送消息（不重复广播）
func (s *MessageService) SendMessageFromWS(ctx context.Context, senderID int64, req *SendMessageRequest) (*model.Message, error) {
	// 验证逻辑与 SendMessage 相同，但不触发广播（因为 WebSocket 会直接发送）
	_, err := s.chatRepo.FindByID(ctx, req.ChatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	_, err = s.chatRepo.GetMember(ctx, req.ChatID, senderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotAuthorized
		}
		return nil, err
	}

	message := &model.Message{
		ChatID:    req.ChatID,
		SenderID:  senderID,
		Type:      req.Type,
		Content:   req.Content,
		MediaURL:  req.MediaURL,
		Duration:  req.Duration,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		ReplyID:   req.ReplyID,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		s.logger.Error("failed to create message from WS", zap.Error(err))
		return nil, err
	}

	return message, nil
}

func (s *MessageService) GetMessages(ctx context.Context, chatID, userID int64, offset, limit int) ([]*model.Message, error) {
	// Check if user is a member of the chat
	_, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotAuthorized
		}
		return nil, err
	}

	messages, err := s.messageRepo.FindByChatID(ctx, chatID, offset, limit)
	if err != nil {
		s.logger.Error("failed to get messages", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, messageID, userID int64) error {
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	// Only sender can delete their own message
	if message.SenderID != userID {
		return ErrNotAuthorized
	}

	return s.messageRepo.Delete(ctx, messageID)
}

func (s *MessageService) GetMessageByID(ctx context.Context, messageID int64) (*model.Message, error) {
	return s.messageRepo.FindByID(ctx, messageID)
}

// AckMessages 批量确认消息已读
// 返回成功标记为已读的消息列表
func (s *MessageService) AckMessages(ctx context.Context, userID, chatID int64, messageIDs []int64) ([]*model.Message, error) {
	// 验证用户是否属于该聊天室
	_, err := s.chatRepo.GetMember(ctx, chatID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotAuthorized
		}
		return nil, err
	}

	// 更新消息为已读状态
	readMessages, err := s.messageRepo.AckMessages(ctx, userID, chatID, messageIDs)
	if err != nil {
		s.logger.Error("failed to ack messages", zap.Error(err))
		return nil, err
	}

	return readMessages, nil
}
