package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/backend/internal/model"
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

// FindBySeqIDsGreaterThan 查询SeqID大于指定值且属于指定聊天室的消息
func (r *MessageRepository) FindBySeqIDsGreaterThan(ctx context.Context, chatIDs []int64, lastSeqID int64, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id IN ? AND seq_id > ? AND is_deleted = ?", chatIDs, lastSeqID, false).
		Order("seq_id ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

// FindReadStatusUpdates 查询用户未读的消息（用于同步已读状态）
func (r *MessageRepository) FindUnreadByChatIDs(ctx context.Context, chatIDs []int64, userID int64) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id IN ? AND sender_id != ? AND is_read = ?", chatIDs, userID, false).
		Order("seq_id ASC").
		Find(&messages).Error
	return messages, err
}

// AckMessages 批量标记消息为已读
// 返回已更新且发送者是其他用户的消息列表
func (r *MessageRepository) AckMessages(ctx context.Context, userID, chatID int64, messageIDs []int64) ([]*model.Message, error) {
	now := r.db.NowFunc()

	// 只标记不是自己发送的消息为已读
	result := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("chat_id = ? AND sender_id != ? AND id IN ? AND is_read = ?", chatID, userID, messageIDs, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return nil, result.Error
	}

	// 查询被更新的消息，返回给调用者用于通知发送者
	var messages []*model.Message
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND sender_id != ? AND id IN ?", chatID, userID, messageIDs).
		Find(&messages).Error

	return messages, err
}
