package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Phone     string         `gorm:"uniqueIndex;size:20" json:"phone"`
	Email     string         `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Nickname  string         `gorm:"size:100" json:"nickname"`
	Avatar    string         `gorm:"size:500" json:"avatar"`
	Bio       string         `gorm:"size:500" json:"bio"`
	Status    int            `gorm:"default:1" json:"status"` // 1: normal, 2: banned
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type Message struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	SeqID      int64          `gorm:"uniqueIndex;not null" json:"seq_id"` // 全局唯一序列号，用于前端去重
	ChatID     int64          `gorm:"index;not null" json:"chat_id"`
	SenderID   int64          `gorm:"index;not null" json:"sender_id"`
	Type       int            `gorm:"type:tinyint;default:1" json:"type"` // 1: text, 2: image, 3: file, 4: voice, 5: location
	Content    string         `gorm:"type:text" json:"content"`
	MediaURL   string         `gorm:"size:500" json:"media_url"`
	Duration   int            `json:"duration"` // for voice
	Latitude   float64        `json:"latitude"` // for location
	Longitude  float64        `json:"longitude"`
	ReplyID    int64          `gorm:"index" json:"reply_id"`
	IsDeleted  bool           `gorm:"default:false" json:"is_deleted"`
	IsRead     bool           `gorm:"default:false" json:"is_read"`     // 消息是否已读
	ReadAt     *time.Time     `json:"read_at"`                         // 消息已读时间
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}

type Chat struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"size:100" json:"name"`
	Type         int            `gorm:"type:tinyint;not null" json:"type"` // 1: private, 2: group, 3: channel
	Avatar       string         `gorm:"size:500" json:"avatar"`
	OwnerID      int64          `gorm:"index" json:"owner_id"`
	MemberCount int            `gorm:"default:0" json:"member_count"`
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Chat) TableName() string {
	return "chats"
}

type ChatMember struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID    int64     `gorm:"index;not null" json:"chat_id"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	Role      int       `gorm:"type:tinyint;default:1" json:"role"` // 1: member, 2: admin, 3: owner
	Nickname  string    `gorm:"size:100" json:"nickname"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ChatMember) TableName() string {
	return "chat_members"
}

type UserSession struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"index;not null" json:"user_id"`
	Token       string    `gorm:"uniqueIndex;size:500;not null" json:"token"`
	DeviceID    string    `gorm:"size:100" json:"device_id"`
	DeviceName  string    `gorm:"size:100" json:"device_name"`
	FCMToken    string    `gorm:"size:500" json:"fcm_token"` // Firebase Cloud Messaging Token
	DeviceType  string    `gorm:"size:20" json:"device_type"` // ios, android, web
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
