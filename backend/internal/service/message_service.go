package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/telegram-go/backend/internal/model"
	"github.com/telegram-go/backend/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrChatNotFound    = errors.New("chat not found")
	ErrMessageNotFound = errors.New("message not found")
	ErrNotAuthorized   = errors.New("not authorized to perform this action")
)

type UserClaims struct {
	UserID int64
}

type MessageService struct {
	messageRepo *repository.MessageRepository
	chatRepo    *repository.ChatRepository
	userRepo    *repository.UserRepository
	logger      *zap.Logger
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

	// Create message
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
		s.logger.Error("failed to create message", zap.Error(err))
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
