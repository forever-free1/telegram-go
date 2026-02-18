package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/internal/model"
	"github.com/forever-free1/telegram-go/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrContactNotFound = errors.New("contact not found")
	ErrContactExists  = errors.New("contact already exists")
)

type ContactService struct {
	userRepository    *repository.UserRepository
	contactRepository *repository.ContactRepository
	logger           *zap.Logger
}

func NewContactService(
	userRepository *repository.UserRepository,
	contactRepository *repository.ContactRepository,
	logger *zap.Logger,
) *ContactService {
	return &ContactService{
		userRepository:    userRepository,
		contactRepository: contactRepository,
		logger:           logger,
	}
}

// ContactInput 联系人输入
type ContactInput struct {
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// ContactWithUser 联系人 + 用户信息
type ContactWithUser struct {
	*model.Contact
	User *model.User `json:"user,omitempty"`
}

// SyncContacts 同步通讯录
// 接收前端上传的手机号列表，在数据库中查找已注册的用户，
// 将他们保存为该用户的联系人，并返回匹配到的用户列表
func (s *ContactService) SyncContacts(ctx context.Context, userID int64, contacts []ContactInput) ([]*ContactWithUser, error) {
	if len(contacts) == 0 {
		return []*ContactWithUser{}, nil
	}

	// 提取手机号列表
	phones := make([]string, 0, len(contacts))
	for _, c := range contacts {
		if c.Phone != "" {
			phones = append(phones, c.Phone)
		}
	}

	if len(phones) == 0 {
		return []*ContactWithUser{}, nil
	}

	// 批量查询已注册的用户
	users, err := s.userRepository.FindByPhones(ctx, phones)
	if err != nil {
		s.logger.Error("failed to find users by phones", zap.Error(err))
		return nil, err
	}

	// 建立手机号到用户的映射
	phoneToUser := make(map[string]*model.User)
	for _, user := range users {
		if user.Phone != "" {
			phoneToUser[user.Phone] = user
		}
	}

	// 收集已匹配的用户ID
	matchedUserIDs := make(map[int64]bool)
	for _, user := range users {
		matchedUserIDs[user.ID] = true
	}

	// 创建或更新联系人关系
	for _, input := range contacts {
		user, exists := phoneToUser[input.Phone]
		if !exists {
			// 用户不存在，跳过
			continue
		}

		// 跳过自己
		if user.ID == userID {
			continue
		}

		// 检查是否已经是联系人
		existing, _ := s.contactRepository.FindContact(ctx, userID, user.ID)
		isMutual := false
		if existing != nil {
			isMutual = existing.IsMutual
		}

		// 如果对方也将我添加为联系人，则互为联系人
		if _, mutual := s.contactRepository.FindContact(ctx, user.ID, userID); mutual == nil {
			isMutual = false
		} else {
			isMutual = true
			// 更新对方的联系人关系为互为联系人
			s.contactRepository.Upsert(ctx, &model.Contact{
				UserID:    user.ID,
				ContactID: userID,
				IsMutual:  true,
			})
		}

		contact := &model.Contact{
			UserID:     userID,
			ContactID:  user.ID,
			Phone:      input.Phone,
			FirstName:  input.FirstName,
			LastName:   input.LastName,
			IsMutual:   isMutual,
		}

		if err := s.contactRepository.Upsert(ctx, contact); err != nil {
			s.logger.Error("failed to upsert contact", zap.Error(err))
			continue
		}
	}

	// 返回匹配到的用户列表
	result := make([]*ContactWithUser, 0, len(users))
	for _, user := range users {
		if user.ID == userID {
			continue
		}
		result = append(result, &ContactWithUser{
			Contact: &model.Contact{
				ContactID: user.ID,
				Phone:     user.Phone,
			},
			User: user,
		})
	}

	return result, nil
}

// GetContacts 获取联系人列表
func (s *ContactService) GetContacts(ctx context.Context, userID int64) ([]*ContactWithUser, error) {
	contacts, err := s.contactRepository.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to find contacts", zap.Error(err))
		return nil, err
	}

	// 获取每个联系人的用户信息
	result := make([]*ContactWithUser, 0, len(contacts))
	for _, contact := range contacts {
		user, err := s.userRepository.FindByID(ctx, contact.ContactID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 用户已被删除
				result = append(result, &ContactWithUser{
					Contact: contact,
					User:    nil,
				})
				continue
			}
			s.logger.Error("failed to find user", zap.Error(err))
			continue
		}

		result = append(result, &ContactWithUser{
			Contact: contact,
			User:    user,
		})
	}

	return result, nil
}

// AddContact 添加联系人
func (s *ContactService) AddContact(ctx context.Context, userID int64, contactID int64) error {
	// 检查联系人是否存在
	_, err := s.userRepository.FindByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// 不能添加自己
	if userID == contactID {
		return errors.New("cannot add yourself as contact")
	}

	// 检查是否已存在
	existing, _ := s.contactRepository.FindContact(ctx, userID, contactID)
	if existing != nil {
		return ErrContactExists
	}

	// 检查对方是否也将我添加为联系人
	isMutual := false
	if _, err := s.contactRepository.FindContact(ctx, contactID, userID); err == nil {
		isMutual = true
	}

	contact := &model.Contact{
		UserID:     userID,
		ContactID:  contactID,
		IsMutual:   isMutual,
	}

	if err := s.contactRepository.Create(ctx, contact); err != nil {
		s.logger.Error("failed to create contact", zap.Error(err))
		return err
	}

	// 如果是互为联系人，更新对方的关系
	if isMutual {
		s.contactRepository.Upsert(ctx, &model.Contact{
			UserID:    contactID,
			ContactID: userID,
			IsMutual:  true,
		})
	}

	return nil
}

// DeleteContact 删除联系人
func (s *ContactService) DeleteContact(ctx context.Context, userID, contactID int64) error {
	// 检查关系是否存在
	existing, err := s.contactRepository.FindContact(ctx, userID, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrContactNotFound
		}
		return err
	}

	// 删除联系人
	if err := s.contactRepository.Delete(ctx, userID, contactID); err != nil {
		s.logger.Error("failed to delete contact", zap.Error(err))
		return err
	}

	// 如果之前是互为联系人，更新对方的关系
	if existing.IsMutual {
		s.contactRepository.Upsert(ctx, &model.Contact{
			UserID:    contactID,
			ContactID: userID,
			IsMutual:  false,
		})
	}

	return nil
}
