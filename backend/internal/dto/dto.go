package dto

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// Request DTOs
type RegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" form:"password" binding:"required,min=6"`
	Phone    string `json:"phone" form:"phone"`
	Email    string `json:"email" form:"email"`
	Nickname string `json:"nickname" form:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type SendMessageRequest struct {
	ChatID    int64   `json:"chat_id" form:"chat_id" binding:"required"`
	Type      int     `json:"type" form:"type" binding:"required,min=1,max=5"`
	Content   string  `json:"content" form:"content"`
	MediaURL  string  `json:"media_url" form:"media_url"`
	Duration  int     `json:"duration" form:"duration"`
	Latitude  float64 `json:"latitude" form:"latitude"`
	Longitude float64 `json:"longitude" form:"longitude"`
	ReplyID   int64   `json:"reply_id" form:"reply_id"`
}

type GetMessagesRequest struct {
	ChatID int64 `json:"chat_id" form:"chat_id" binding:"required"`
	Offset  int   `json:"offset" form:"offset"`
	Limit   int   `json:"limit" form:"limit,default=50"`
}

type CreateChatRequest struct {
	Name string `json:"name" form:"name"`
	Type int    `json:"type" form:"type" binding:"required,min=1,max=3"`
}

type AddChatMemberRequest struct {
	ChatID int64 `json:"chat_id" form:"chat_id" binding:"required"`
	UserID int64 `json:"user_id" form:"user_id" binding:"required"`
}
