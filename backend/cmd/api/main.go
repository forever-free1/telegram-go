package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/telegram-go/backend/internal/config"
	"github.com/telegram-go/backend/internal/database"
	"github.com/telegram-go/backend/internal/handler"
	"github.com/telegram-go/backend/internal/middleware"
	"github.com/telegram-go/backend/internal/repository"
	"github.com/telegram-go/backend/internal/service"
	"github.com/telegram-go/backend/internal/websocket"
	"go.uber.org/zap"
)

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

	// Setup handlers
	authHandler := handler.NewAuthHandler(authService)
	messageHandler := handler.NewMessageHandler(messageService)
	chatHandler := handler.NewChatHandler(chatService)

	// Setup WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Setup router
	router := gin.Default()

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
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Starting server", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
