package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/telegram-go/backend/internal/model"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *model.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*model.UserSession, error) {
	var session model.UserSession
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.UserSession{}).Error
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
}

func (r *SessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]*model.UserSession, error) {
	var sessions []*model.UserSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}
