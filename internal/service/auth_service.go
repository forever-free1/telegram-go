package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/forever-free1/telegram-go/internal/config"
	"github.com/forever-free1/telegram-go/internal/model"
	"github.com/forever-free1/telegram-go/internal/repository"
	"github.com/forever-free1/telegram-go/pkg/crypto"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidToken      = errors.New("invalid token")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtConfig   *config.JWTConfig
	logger      *zap.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	jwtConfig *config.JWTConfig,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtConfig:   jwtConfig,
		logger:      logger,
	}
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, _ := s.userRepo.FindByUsername(ctx, req.Username)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", zap.Error(err))
		return nil, err
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Phone:    req.Phone,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &model.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpireHour)),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check password
	if !crypto.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidPassword
	}

	// Generate token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &model.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpireHour)),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		s.logger.Error("failed to create session", zap.Error(err))
	}

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*model.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID := int64(claims["user_id"].(float64))
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, token)
}

func (s *AuthService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpireHour)).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}
