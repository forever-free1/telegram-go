//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/forever-free1/telegram-go/internal/config"
	"github.com/forever-free1/telegram-go/internal/database"
	"github.com/forever-free1/telegram-go/internal/handler"
	"github.com/forever-free1/telegram-go/internal/middleware"
	"github.com/forever-free1/telegram-go/internal/repository"
	"github.com/forever-free1/telegram-go/internal/service"
	"github.com/forever-free1/telegram-go/pkg/snowflake"
	"go.uber.org/zap"
)

func InitializeApp(cfgPath string) (*App, error) {
	wire.Build(
		// Config
		config.LoadConfig,

		// Database
		database.NewDatabase,

		// Repositories
		repository.NewUserRepository,
		repository.NewMessageRepository,
		repository.NewChatRepository,
		repository.NewSessionRepository,

		// Services
		service.NewAuthService,
		service.NewMessageService,

		// Handlers
		handler.NewAuthHandler,
		handler.NewMessageHandler,

		// Middleware
		middleware.AuthMiddleware,

		// Utils
		snowflake.NewSnowflake,

		// Logger
		zap.NewDevelopment,

		// App
		NewApp,
	)
	return nil, nil
}

type App struct {
	AuthHandler    *handler.AuthHandler
	MessageHandler *handler.MessageHandler
	AuthMiddleware func() gin.HandlerFunc
	Config         *config.Config
	Logger         *zap.Logger
}
