package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/telegram-go/backend/internal/model"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(ctx context.Context, chat *model.Chat) error {
	return r.db.WithContext(ctx).Create(chat).Error
}

func (r *ChatRepository) FindByID(ctx context.Context, id int64) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *ChatRepository) Update(ctx context.Context, chat *model.Chat) error {
	return r.db.WithContext(ctx).Save(chat).Error
}

func (r *ChatRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Chat{}, id).Error
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID int64) ([]*model.Chat, error) {
	var chats []*model.Chat
	err := r.db.WithContext(ctx).
		Joins("JOIN chat_members ON chat_members.chat_id = chats.id").
		Where("chat_members.user_id = ?", userID).
		Find(&chats).Error
	return chats, err
}

// ChatMember methods
func (r *ChatRepository) AddMember(ctx context.Context, member *model.ChatMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *ChatRepository) RemoveMember(ctx context.Context, chatID, userID int64) error {
	return r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&model.ChatMember{}).Error
}

func (r *ChatRepository) GetMember(ctx context.Context, chatID, userID int64) (*model.ChatMember, error) {
	var member model.ChatMember
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *ChatRepository) GetMembers(ctx context.Context, chatID int64) ([]*model.ChatMember, error) {
	var members []*model.ChatMember
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Find(&members).Error
	return members, err
}
