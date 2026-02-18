package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/forever-free1/telegram-go/internal/config"
	"github.com/forever-free1/telegram-go/internal/database"
	"github.com/forever-free1/telegram-go/internal/handler"
	"github.com/forever-free1/telegram-go/internal/middleware"
	"github.com/forever-free1/telegram-go/internal/model"
	"github.com/forever-free1/telegram-go/internal/repository"
	"github.com/forever-free1/telegram-go/internal/service"
	"github.com/forever-free1/telegram-go/internal/websocket"
	"go.uber.org/zap"
)

// wsMessageHandler WebSocket 消息处理器
// 负责将 WebSocket 接收到的消息保存到数据库
type wsMessageHandler struct {
	messageService *service.MessageService
	hub           *websocket.Hub
}

// OnMessageSaved 实现 websocket.MessageEventHandler 接口
func (h *wsMessageHandler) OnMessageSaved(message *model.Message) {
	// 这里实际上不会直接调用，因为 WebSocket 消息由 readPump 处理
	// 但为了实现接口，需要有这个方法
	// WebSocket 消息会在 readPump 中保存后直接通过 hub.broadcast 发送
}

func main() {
	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Setup database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	chatRepo := repository.NewChatRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Setup services
	authService := service.NewAuthService(userRepo, sessionRepo, &cfg.JWT, logger)
	messageService := service.NewMessageService(messageRepo, chatRepo, userRepo, logger)
	chatService := service.NewChatService(chatRepo, userRepo, logger)
	fileService := service.NewFileService(cfg.Upload.Path, cfg.Upload.BaseURL, logger, service.WithMaxSize(cfg.Upload.MaxSize))
	notificationService := service.NewNotificationService(logger)

	// Setup WebSocket hub (must be created before handlers)
	wsHub := websocket.NewHub()

	// 设置 Hub 的在线用户检查回调
	wsHub.SetOnlineChecker(func(userID int64) bool {
		return notificationService.IsUserOnline(userID)
	})

	// Setup handlers
	authHandler := handler.NewAuthHandler(authService)
	messageHandler := handler.NewMessageHandler(messageService, wsHub)
	chatHandler := handler.NewChatHandler(chatService)
	uploadHandler := handler.NewUploadHandler(fileService)
	deviceHandler := handler.NewDeviceHandler(notificationService)

	// 设置消息服务使用离线推送
	messageService.SetPushService(notificationService)

	// 设置消息广播器：MessageService -> Hub
	// 当 REST API 发送消息时，保存成功后通过 Hub 广播
	messageService.SetBroadcaster(service.MessageBroadcasterFunc(func(msg *model.Message) {
		wsHub.OnMessageSaved(msg)
	}))

	// 设置 WebSocket 消息处理器：Hub -> MessageService
	// 当 WebSocket 收到消息时，保存到数据库
	wsHub.SetMessageHandler(&wsMessageHandler{
		messageService: messageService,
		hub:           wsHub,
	})

	// 设置消息保存回调
	// 当 WebSocket 收到聊天消息时，保存到数据库
	wsHub.SetMessageSaver(func(ctx context.Context, msg *websocket.WSMessage) (*model.Message, error) {
		req := &service.SendMessageRequest{
			ChatID:   msg.ChatID,
			Type:     msg.MsgType,
			Content:  msg.Content,
			MediaURL: msg.MediaURL,
		}
		return messageService.SendMessageFromWS(ctx, msg.SenderID, req)
	})

	go wsHub.Run()

	// Setup router
	router := gin.Default()

	// 静态文件服务 - 上传的文件
	router.Static("/static", cfg.Upload.Path)

	// Public routes
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)

	// WebSocket endpoint
	router.GET("/ws", websocket.ServeWS(wsHub))

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// User routes
		protected.GET("/user/me", authHandler.GetCurrentUser)
		protected.POST("/auth/logout", authHandler.Logout)

		// Upload routes
		protected.POST("/upload", uploadHandler.Upload)

		// Device routes
		protected.POST("/device/token", deviceHandler.RegisterToken)
		protected.DELETE("/device/token", deviceHandler.UnregisterToken)

		// Chat routes
		protected.POST("/chats", chatHandler.CreateChat)
		protected.GET("/chats", chatHandler.GetUserChats)
		protected.GET("/chats/:id", chatHandler.GetChat)
		protected.POST("/chats/members", chatHandler.AddMember)
		protected.DELETE("/chats/members", chatHandler.RemoveMember)
		protected.GET("/chats/:id/members", chatHandler.GetMembers)

		// Message routes
		protected.POST("/messages", messageHandler.SendMessage)
		protected.GET("/messages", messageHandler.GetMessages)
		protected.DELETE("/messages/:id", messageHandler.DeleteMessage)
		protected.POST("/messages/ack", messageHandler.AckMessage)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Starting server", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
