package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/internal/model"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create 创建联系人
func (r *ContactRepository) Create(ctx context.Context, contact *model.Contact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

// Upsert  upsert 联系人（如果已存在则更新）
func (r *ContactRepository) Upsert(ctx context.Context, contact *model.Contact) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
		Assign(map[string]interface{}{
			"phone":      contact.Phone,
			"first_name": contact.FirstName,
			"last_name":  contact.LastName,
			"is_mutual":  contact.IsMutual,
		}).
		FirstOrCreate(contact).Error
}

// Delete 删除联系人
func (r *ContactRepository) Delete(ctx context.Context, userID, contactID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		Delete(&model.Contact{}).Error
}

// FindByUserID 获取用户的所有联系人
func (r *ContactRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.Contact, error) {
	var contacts []*model.Contact
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&contacts).Error
	return contacts, err
}

// FindContact 查询特定联系人关系
func (r *ContactRepository) FindContact(ctx context.Context, userID, contactID int64) (*model.Contact, error) {
	var contact model.Contact
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", userID, contactID).
		First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// FindByPhone 通过手机号查找联系人
func (r *ContactRepository) FindByPhone(ctx context.Context, userID int64, phone string) (*model.Contact, error) {
	var contact model.Contact
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND phone = ?", userID, phone).
		First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// GetContactUsers 获取联系人的用户信息
func (r *ContactRepository) GetContactUsers(ctx context.Context, userID int64) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).
		Table("contacts").
		Select("users.*").
		Joins("INNER JOIN users ON users.id = contacts.contact_id").
		Where("contacts.user_id = ?", userID).
		Find(&users).Error
	return users, err
}

// BatchCreate 批量创建联系人
func (r *ContactRepository) BatchCreate(ctx context.Context, contacts []*model.Contact) error {
	if len(contacts) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(contacts).Error
}

// DeleteAll 删除用户的所有联系人
func (r *ContactRepository) DeleteAll(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.Contact{}).Error
}
