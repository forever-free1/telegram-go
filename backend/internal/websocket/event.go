package websocket

import (
	"github.com/forever-free1/telegram-go/backend/internal/model"
)

// MessageEventHandler 消息事件处理器接口
// 用于解决 WebSocket Hub 和 MessageService 之间的循环依赖
type MessageEventHandler interface {
	// OnMessageSaved 消息保存成功后触发，用于广播给聊天室成员
	OnMessageSaved(message *model.Message)
}

// Hub 实现了 MessageEventHandler 接口
var _ MessageEventHandler = (*Hub)(nil)
