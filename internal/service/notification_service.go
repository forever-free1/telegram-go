package service

import (
	"context"
	"log"
	"sync"

	"go.uber.org/zap"
)

// PushService 推送服务接口
type PushService interface {
	// Push 发送离线推送
	Push(ctx context.Context, userID int64, title, content string, data map[string]string) error

	// PushToDevice 发送离线推送到指定设备
	PushToDevice(ctx context.Context, fcmToken, title, content string, data map[string]string) error

	// RegisterToken 注册设备 Token
	RegisterToken(ctx context.Context, userID int64, fcmToken, deviceType string) error

	// UnregisterToken 注销设备 Token
	UnregisterToken(ctx context.Context, userID int64, fcmToken string) error

	// IsUserOnline 检查用户是否在线
	IsUserOnline(userID int64) bool
}

// NotificationService 离线消息推送服务
// 实现 PushService 接口，提供离线推送功能
type NotificationService struct {
	logger *zap.Logger
	// 这里可以注入 FCM 客户端或其他推送服务
	fcmClient interface{}
	// 模拟在线用户检查
	onlineUsers map[int64]bool
	mu         sync.RWMutex
}

// NewNotificationService 创建推送服务
func NewNotificationService(logger *zap.Logger) *NotificationService {
	return &NotificationService{
		logger:     logger,
		onlineUsers: make(map[int64]bool),
	}
}

// Push 发送离线推送
// 如果用户在线则不发送（由 WebSocket 处理）
func (s *NotificationService) Push(ctx context.Context, userID int64, title, content string, data map[string]string) error {
	// TODO: 实现实际的 FCM 推送
	// 这里先实现 Mock 逻辑

	// 检查用户是否在线
	if s.IsUserOnline(userID) {
		s.logger.Debug("user is online, skip push", zap.Int64("user_id", userID))
		return nil
	}

	// 用户离线，发送推送
	s.logger.Info("sending offline push",
		zap.Int64("user_id", userID),
		zap.String("title", title),
		zap.String("content", content))

	// Mock: 打印推送信息
	log.Printf("[MOCK PUSH] to user %d: %s - %s", userID, title, content)

	// TODO: 实际实现
	// 1. 从数据库获取用户的 FCM Token
	// 2. 调用 FCM API 发送推送
	// 3. 处理推送结果

	return nil
}

// PushToDevice 发送离线推送到指定设备
func (s *NotificationService) PushToDevice(ctx context.Context, fcmToken, title, content string, data map[string]string) error {
	// TODO: 实现实际的 FCM 推送到指定设备

	s.logger.Info("sending push to device",
		zap.String("fcm_token", fcmToken[:min(10, len(fcmToken))]+"..."),
		zap.String("title", title),
		zap.String("content", content))

	// Mock: 打印推送信息
	log.Printf("[MOCK PUSH] to device: %s - %s", title, content)

	return nil
}

// RegisterToken 注册设备 Token
func (s *NotificationService) RegisterToken(ctx context.Context, userID int64, fcmToken, deviceType string) error {
	// TODO: 将 FCM Token 存储到数据库 UserSession 表

	s.logger.Info("registering device token",
		zap.Int64("user_id", userID),
		zap.String("device_type", deviceType),
		zap.String("fcm_token", fcmToken[:min(10, len(fcmToken))]+"..."))

	return nil
}

// UnregisterToken 注销设备 Token
func (s *NotificationService) UnregisterToken(ctx context.Context, userID int64, fcmToken string) error {
	// TODO: 从数据库删除 FCM Token

	s.logger.Info("unregistering device token",
		zap.Int64("user_id", userID),
		zap.String("fcm_token", fcmToken[:min(10, len(fcmToken))]+"..."))

	return nil
}

// SetUserOnline 设置用户在线状态
func (s *NotificationService) SetUserOnline(userID int64, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onlineUsers[userID] = online
}

// IsUserOnline 检查用户是否在线
func (s *NotificationService) IsUserOnline(userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.onlineUsers[userID]
}

// NotificationPayload 推送消息体
type NotificationPayload struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	Sound    string            `json:"sound,omitempty"`
	Badge    int               `json:"badge,omitempty"`
	ClickAction string         `json:"click_action,omitempty"`
}

// APNsPayload APNs 专用负载
type APNsPayload struct {
	Aps *APNsAps `json:"aps"`
}

// APNsAps APNs 消息结构
type APNsAps struct {
	Alert            *APNsAlert `json:"alert,omitempty"`
	Badge            int        `json:"badge,omitempty"`
	Sound            string     `json:"sound,omitempty"`
	ContentAvailable int        `json:"content-available,omitempty"`
}

// APNsAlert APNs 提示
type APNsAlert struct {
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
}

// FCMMessage FCM 消息结构
type FCMMessage struct {
	Token        string                 `json:"token,omitempty"`
	Topic        string                 `json:"topic,omitempty"`
	Condition    string                 `json:"condition,omitempty"`
	Data         map[string]string      `json:"data,omitempty"`
	Notification *NotificationPayload   `json:"notification,omitempty"`
	Android      *AndroidConfig         `json:"android,omitempty"`
	APNS         *APNsPayload           `json:"apns,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	TimeToLive   int                    `json:"ttl,omitempty"`
}

// AndroidConfig Android 配置
type AndroidConfig struct {
	Priority        string            `json:"priority,omitempty"`
	TTL             string            `json:"ttl,omitempty"`
	RestrictedPackageName string       `json:"restricted_package_name,omitempty"`
	Notification    *AndroidNotification `json:"notification,omitempty"`
}

// AndroidNotification Android 通知
type AndroidNotification struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction string `json:"click_action,omitempty"`
}
