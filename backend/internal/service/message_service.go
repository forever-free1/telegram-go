package service

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"gorm.io/gorm"

	"github.com/forever-free1/telegram-go/backend/internal/model"
	"github.com/forever-free1/telegram-go/backend/internal/repository"
	"github.com/forever-free1/telegram-go/backend/pkg/snowflake"
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
	pushService  PushService // 离线推送服务
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

// SetPushService 设置离线推送服务
func (s *MessageService) SetPushService(pushService PushService) {
	s.pushService = pushService
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
	chat, err := s.chatRepo.FindByID(ctx, req.ChatID)
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
		SeqID:     snowflake.GenerateID(),
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

	// 消息保存成功后，触发 WebSocket 广播
	s.broadcast(message)

	// 发送离线推送（如果是私聊）
	// 群聊的离线推送逻辑更复杂，这里先实现私聊
	if chat.Type == 1 { // private chat
		go s.sendOfflinePush(ctx, message, senderID)
	}

	return message, nil
}

// sendOfflinePush 发送离线推送
// 检查聊天室成员是否在线，不在线则发送推送
func (s *MessageService) sendOfflinePush(ctx context.Context, message *model.Message, senderID int64) {
	if s.pushService == nil {
		return
	}

	// 获取聊天室成员
	members, err := s.chatRepo.GetMembers(ctx, message.ChatID)
	if err != nil {
		s.logger.Error("failed to get chat members for offline push", zap.Error(err))
		return
	}

	// 大群限制：只对前100名成员发送离线推送
	// 避免推送风暴
	const maxPushRecipients = 100
	if len(members) > maxPushRecipients {
		members = members[:maxPushRecipients]
	}

	// 获取发送者信息
	sender, err := s.userRepo.FindByID(ctx, senderID)
	if err != nil {
		s.logger.Error("failed to get sender for offline push", zap.Error(err))
		return
	}

	// 构建推送内容
	title := sender.Nickname
	if title == "" {
		title = sender.Username
	}

	var content string
	switch message.Type {
	case 1: // text
		content = message.Content
	case 2: // image
		content = "[图片]"
	case 3: // file
		content = "[文件]"
	case 4: // voice
		content = "[语音]"
	case 5: // location
		content = "[位置]"
	}

	// 限制内容长度
	if len(content) > 100 {
		content = content[:100] + "..."
	}

	// 为每个离线成员异步发送推送
	for _, member := range members {
		// 跳过发送者自己
		if member.UserID == senderID {
			continue
		}

		// 异步检查并发送推送
		go func(memberUserID int64) {
			// 检查用户是否在线（通过推送服务检查）
			isOnline := s.pushService.IsUserOnline(memberUserID)
			if !isOnline {
				// 用户离线，发送推送
				data := map[string]string{
					"chat_id":    strconv.FormatInt(message.ChatID, 10),
					"message_id": strconv.FormatInt(message.ID, 10),
					"sender_id":  strconv.FormatInt(message.SenderID, 10),
				}

				err := s.pushService.Push(ctx, memberUserID, title, content, data)
				if err != nil {
					s.logger.Error("failed to send offline push",
						zap.Int64("user_id", memberUserID),
						zap.Error(err))
				}
			}
		}(member.UserID)
	}
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
		SeqID:     snowflake.GenerateID(),
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

// SyncRequest 增量同步请求
type SyncRequest struct {
	LastSeqID int64 `json:"last_seq_id"`
}

// SyncResponse 增量同步响应
type SyncResponse struct {
	Messages   []*model.Message `json:"messages"`
	ReadAcks   []*model.Message `json:"read_acks"` // 已读确认的消息
	SeqID      int64            `json:"seq_id"`     // 当前最新SeqID
	HasMore    bool             `json:"has_more"`  // 告诉前端是否还有更多消息
}

// Sync 增量同步 - 获取lastSeqID之后的所有消息
func (s *MessageService) Sync(ctx context.Context, userID int64, lastSeqID int64) (*SyncResponse, error) {
	chatIDs, err := s.chatRepo.GetChatIDsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user chat IDs", zap.Error(err))
		return nil, err
	}

	if len(chatIDs) == 0 {
		return &SyncResponse{
			Messages: []*model.Message{},
			ReadAcks: []*model.Message{},
			SeqID:    lastSeqID,
			HasMore:  false,
		}, nil
	}

	// 强制限制一次最多拉取 500 条
	limit := 500
	messages, err := s.messageRepo.FindBySeqIDsGreaterThan(ctx, chatIDs, lastSeqID, limit)
	if err != nil {
		s.logger.Error("failed to sync messages", zap.Error(err))
		return nil, err
	}

	currentSeqID := lastSeqID
	for _, msg := range messages {
		if msg.SeqID > currentSeqID {
			currentSeqID = msg.SeqID
		}
	}

	// 如果拉回来的消息数量等于 limit，说明数据库里很可能还有没拉完的数据
	hasMore := len(messages) == limit

	return &SyncResponse{
		Messages: messages,
		ReadAcks: []*model.Message{}, // 暂不实现已读同步
		SeqID:    currentSeqID,
		HasMore:  hasMore,
	}, nil
}
