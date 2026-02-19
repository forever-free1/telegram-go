package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/backend/internal/model"
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
	// 先查询是否存在
	var existing model.Contact
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
		First(&existing).Error

	if err == nil {
		// 已存在：如果是互为联系人，不要覆盖为非互为
		if existing.IsMutual && !contact.IsMutual {
			contact.IsMutual = true
		}
	}

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

// FindByUserIDs 批量查询指定用户列表的联系人关系
// key: userID, value: map of contactID -> Contact
func (r *ContactRepository) FindByUserIDs(ctx context.Context, userIDs []int64) (map[int64]map[int64]*model.Contact, error) {
	if len(userIDs) == 0 {
		return make(map[int64]map[int64]*model.Contact), nil
	}

	var contacts []*model.Contact
	err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Find(&contacts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]map[int64]*model.Contact)
	for _, c := range contacts {
		if result[c.UserID] == nil {
			result[c.UserID] = make(map[int64]*model.Contact)
		}
		result[c.UserID][c.ContactID] = c
	}
	return result, nil
}

// FindReverseContacts 查找将指定用户添加为联系人的所有反向关系
// 即: 哪些用户把我添加为联系人
func (r *ContactRepository) FindReverseContacts(ctx context.Context, userID int64) (map[int64]*model.Contact, error) {
	var contacts []*model.Contact
	err := r.db.WithContext(ctx).
		Where("contact_id = ?", userID).
		Find(&contacts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.Contact)
	for _, c := range contacts {
		result[c.UserID] = c
	}
	return result, nil
}

// BatchUpsert 批量 upsert 联系人
func (r *ContactRepository) BatchUpsert(ctx context.Context, contacts []*model.Contact) error {
	if len(contacts) == 0 {
		return nil
	}

	for _, contact := range contacts {
		// 先查询是否存在
		var existing model.Contact
		err := r.db.WithContext(ctx).
			Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
			First(&existing).Error

		if err == nil {
			// 已存在：如果是互为联系人，不要覆盖为非互为
			if existing.IsMutual && !contact.IsMutual {
				contact.IsMutual = true
			}
		}

		err = r.db.WithContext(ctx).
			Where("user_id = ? AND contact_id = ?", contact.UserID, contact.ContactID).
			Assign(map[string]interface{}{
				"phone":      contact.Phone,
				"first_name": contact.FirstName,
				"last_name":  contact.LastName,
				"is_mutual":  contact.IsMutual,
			}).
			FirstOrCreate(contact).Error
		if err != nil {
			return err
		}
	}
	return nil
}
