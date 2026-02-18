package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/internal/model"
	"github.com/forever-free1/telegram-go/internal/repository"
	"go.uber.org/zap"
)

type ChatService struct {
	chatRepo *repository.ChatRepository
	userRepo *repository.UserRepository
	logger   *zap.Logger
}

func NewChatService(
	chatRepo *repository.ChatRepository,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		userRepo: userRepo,
		logger:   logger,
	}
}

type CreateChatRequest struct {
	Name string `json:"name"`
	Type int    `json:"type" validate:"required,min=1,max=3"` // 1: private, 2: group, 3: channel
}

func (s *ChatService) CreateChat(ctx context.Context, ownerID int64, req *CreateChatRequest) (*model.Chat, error) {
	chat := &model.Chat{
		Name:    req.Name,
		Type:    req.Type,
		OwnerID: ownerID,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		s.logger.Error("failed to create chat", zap.Error(err))
		return nil, err
	}

	// Add owner as member
	member := &model.ChatMember{
		ChatID: chat.ID,
		UserID: ownerID,
		Role:   3, // owner
	}
	if err := s.chatRepo.AddMember(ctx, member); err != nil {
		s.logger.Error("failed to add member", zap.Error(err))
	}

	return chat, nil
}

func (s *ChatService) GetChat(ctx context.Context, chatID int64) (*model.Chat, error) {
	return s.chatRepo.FindByID(ctx, chatID)
}

func (s *ChatService) GetUserChats(ctx context.Context, userID int64) ([]*model.Chat, error) {
	return s.chatRepo.GetUserChats(ctx, userID)
}

func (s *ChatService) AddMember(ctx context.Context, chatID, userID int64, role int) error {
	// Check if user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	member := &model.ChatMember{
		ChatID:   chatID,
		UserID:   userID,
		Role:     role,
		JoinedAt: time.Now(),
	}
	return s.chatRepo.AddMember(ctx, member)
}

func (s *ChatService) RemoveMember(ctx context.Context, chatID, userID int64) error {
	return s.chatRepo.RemoveMember(ctx, chatID, userID)
}

func (s *ChatService) GetMembers(ctx context.Context, chatID int64) ([]*model.ChatMember, error) {
	return s.chatRepo.GetMembers(ctx, chatID)
}

func (s *ChatService) GetOrCreatePrivateChat(ctx context.Context, userID1, userID2 int64) (*model.Chat, error) {
	// This is a simplified implementation - in production you'd want to check
	// if a private chat already exists between these two users
	chats, err := s.chatRepo.GetUserChats(ctx, userID1)
	if err != nil {
		return nil, err
	}

	for _, chat := range chats {
		if chat.Type == 1 { // private chat
			members, _ := s.chatRepo.GetMembers(ctx, chat.ID)
			if len(members) == 2 {
				for _, m := range members {
					if m.UserID == userID2 {
						return chat, nil
					}
				}
			}
		}
	}

	// Create new private chat
	chat := &model.Chat{
		Type:    1,
		OwnerID: userID1,
	}
	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	// Add both users as members
	s.chatRepo.AddMember(ctx, &model.ChatMember{ChatID: chat.ID, UserID: userID1, Role: 3})
	s.chatRepo.AddMember(ctx, &model.ChatMember{ChatID: chat.ID, UserID: userID2, Role: 1})

	return chat, nil
}
