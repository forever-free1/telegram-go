package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/telegram-go/backend/internal/model"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *MessageRepository) FindByID(ctx context.Context, id int64) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) FindByChatID(ctx context.Context, chatID int64, offset, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND is_deleted = ?", chatID, false).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Message{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *MessageRepository) Update(ctx context.Context, message *model.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}
