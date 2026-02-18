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

// SendMessageRequest 发送消息请求
// 消息类型 (Type):
//   - 1: 文本消息 (text)
//   - 2: 图片消息 (image) - 需要先上传图片，获取 MediaURL
//   - 3: 文件消息 (file) - 需要先上传文件，获取 MediaURL
//   - 4: 语音消息 (voice) - 需要先上传音频，获取 MediaURL，可设置 Duration
//   - 5: 位置消息 (location) - 需要设置 Latitude 和 Longitude
//
// 发送媒体消息流程:
//   1. 调用 POST /api/upload 上传文件，获取 URL
//   2. 调用 POST /api/messages 发送消息，Type=2/3/4，MediaURL 填写上一步获取的 URL
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
