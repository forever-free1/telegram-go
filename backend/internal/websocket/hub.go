package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/telegram-go/backend/internal/model"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// MessageSaveHandler 消息保存处理函数类型
type MessageSaveHandler func(ctx context.Context, msg *WSMessage) (*model.Message, error)

// OnlineChecker 用户在线检查函数类型
type OnlineChecker func(userID int64) bool

// Hub WebSocket 消息中心
type Hub struct {
	clients         map[int64]*Client
	register       chan *Client
	unregister     chan *Client
	broadcast      chan *WSMessage
	messageHandler MessageEventHandler    // 消息事件处理器，用于保存消息到数据库
	messageSaver   MessageSaveHandler     // 消息保存回调，用于将 WebSocket 消息保存到数据库
	onlineChecker  OnlineChecker         // 用户在线检查回调
	chatMembers    map[int64]map[int64]bool // chatID -> map[userID] -> joined
	chatMembersMu  sync.RWMutex
	mu             sync.RWMutex
}

// Client WebSocket 客户端连接
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	chatID int64 // 当前所在的聊天室
}

// WSMessage WebSocket 消息结构
// 消息类型 (Type):
//   - "message": 普通消息
//   - "join_chat": 加入聊天室
//   - "leave_chat": 离开聊天室
//   - "WS_MSG_READ": 消息已读回执
//   - "WS_TYPING": 正在输入状态
type WSMessage struct {
	Type       string          `json:"type"`
	SeqID      int64           `json:"seq_id,omitempty"`      // 消息序列号，用于前端去重
	MessageID  int64           `json:"message_id,omitempty"`  // 数据库消息ID
	ChatID     int64           `json:"chat_id,omitempty"`
	SenderID   int64           `json:"sender_id,omitempty"`
	Content    string          `json:"content,omitempty"`
	MediaURL   string          `json:"media_url,omitempty"`
	MsgType    int             `json:"msg_type,omitempty"`    // 消息类型：1:text, 2:image, 3:file, 4:voice, 5:location
	Timestamp  time.Time       `json:"timestamp"`
	Data       json.RawMessage `json:"data,omitempty"`
	MessageIDs []int64         `json:"message_ids,omitempty"` // 用于已读确认
}

// WSMessageType constants
const (
	WSMessageType         = "message"
	WSMsgReadType         = "WS_MSG_READ"
	WSTypingType          = "WS_TYPING"
)

// NewHub 创建新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[int64]*Client),
		register:     make(chan *Client, 10),
		unregister:   make(chan *Client, 10),
		broadcast:    make(chan *WSMessage, 256),
		chatMembers:  make(map[int64]map[int64]bool),
	}
}

// SetMessageHandler 设置消息事件处理器
func (h *Hub) SetMessageHandler(handler MessageEventHandler) {
	h.messageHandler = handler
}

// SetMessageSaver 设置消息保存回调
// 用于在 WebSocket 收到消息时保存到数据库
func (h *Hub) SetMessageSaver(saver MessageSaveHandler) {
	h.messageSaver = saver
}

// SetOnlineChecker 设置用户在线检查回调
// 用于检查用户是否有 WebSocket 连接
func (h *Hub) SetOnlineChecker(checker OnlineChecker) {
	h.onlineChecker = checker
}

// IsUserOnline 检查用户是否有 WebSocket 连接
func (h *Hub) IsUserOnline(userID int64) bool {
	// 首先检查本地连接
	h.mu.RLock()
	_, hasLocalConn := h.clients[userID]
	h.mu.RUnlock()

	if hasLocalConn {
		return true
	}

	// 如果设置了外部在线检查回调，也检查
	if h.onlineChecker != nil {
		return h.onlineChecker(userID)
	}

	return false
}

// Run 启动 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.userID] = client
			// 将用户加入聊天室
			if client.chatID > 0 {
				h.addUserToChat(client.chatID, client.userID)
			}
			h.mu.Unlock()
			log.Printf("User %d connected, total clients: %d", client.userID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				// 将用户从聊天室移除
				if client.chatID > 0 {
					h.removeUserFromChat(client.chatID, client.userID)
				}
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("User %d disconnected, total clients: %d", client.userID, len(h.clients))

		case message := <-h.broadcast:
			h.broadcastToChat(message)
		}
	}
}

func (h *Hub) addUserToChat(chatID, userID int64) {
	h.chatMembersMu.Lock()
	defer h.chatMembersMu.Unlock()

	if h.chatMembers[chatID] == nil {
		h.chatMembers[chatID] = make(map[int64]bool)
	}
	h.chatMembers[chatID][userID] = true
}

func (h *Hub) removeUserFromChat(chatID, userID int64) {
	h.chatMembersMu.Lock()
	defer h.chatMembersMu.Unlock()

	if h.chatMembers[chatID] != nil {
		delete(h.chatMembers[chatID], userID)
	}
}

func (h *Hub) broadcastToChat(msg *WSMessage) {
	h.chatMembersMu.RLock()
	members, ok := h.chatMembers[msg.ChatID]
	h.chatMembersMu.RUnlock()

	if !ok {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	data := h.encodeMessage(msg)
	for userID := range members {
		if client, exists := h.clients[userID]; exists {
			select {
			case client.send <- data:
			default:
				// 发送失败，关闭连接
				close(client.send)
				delete(h.clients, userID)
			}
		}
	}
}

// broadcastToChatExcludeSender 广播消息给聊天室成员，排除发送者
func (h *Hub) broadcastToChatExcludeSender(msg *WSMessage, excludeUserID int64) {
	h.chatMembersMu.RLock()
	members, ok := h.chatMembers[msg.ChatID]
	h.chatMembersMu.RUnlock()

	if !ok {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	data := h.encodeMessage(msg)
	for userID := range members {
		if userID == excludeUserID {
			continue
		}
		if client, exists := h.clients[userID]; exists {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, userID)
			}
		}
	}
}

func (h *Hub) encodeMessage(msg *WSMessage) []byte {
	data, _ := json.Marshal(msg)
	return data
}

// GetClient 获取用户客户端连接
func (h *Hub) GetClient(userID int64) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID int64, msg *WSMessage) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()
	if ok {
		select {
		case client.send <- h.encodeMessage(msg):
		default:
		}
	}
}

// OnMessageSaved 实现 MessageEventHandler 接口
// 消息保存成功后调用此方法进行广播
func (h *Hub) OnMessageSaved(message *model.Message) {
	wsMsg := &WSMessage{
		Type:      "message",
		SeqID:     message.SeqID,
		MessageID: message.ID,
		ChatID:    message.ChatID,
		SenderID:  message.SenderID,
		Content:   message.Content,
		MediaURL:  message.MediaURL,
		MsgType:   message.Type,
		Timestamp: message.CreatedAt,
	}

	select {
	case h.broadcast <- wsMsg:
	default:
		log.Println("broadcast channel is full, message may be dropped")
	}
}

// JoinChat 用户加入聊天室
func (h *Hub) JoinChat(userID, chatID int64) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		// 从旧聊天室移除
		if client.chatID > 0 {
			h.removeUserFromChat(client.chatID, userID)
		}
		// 加入新聊天室
		client.chatID = chatID
		h.addUserToChat(chatID, userID)
	}
}

// LeaveChat 用户离开聊天室
func (h *Hub) LeaveChat(userID, chatID int64) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok && client.chatID == chatID {
		h.removeUserFromChat(chatID, userID)
		client.chatID = 0
	}
}

// ServeWS WebSocket 处理函数
func ServeWS(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}

		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			userID: userID,
		}

		hub.register <- client

		go client.writePump()
		go client.readPump()
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		wsMsg.SenderID = c.userID
		wsMsg.Timestamp = time.Now()

		// 处理加入聊天室消息
		if wsMsg.Type == "join_chat" {
			c.hub.JoinChat(c.userID, wsMsg.ChatID)
			continue
		}

		// 处理离开聊天室消息
		if wsMsg.Type == "leave_chat" {
			c.hub.LeaveChat(c.userID, wsMsg.ChatID)
			continue
		}

		// 处理聊天消息
		if wsMsg.Type == WSMessageType {
			// 如果有消息保存回调，先保存到数据库
			if c.hub.messageSaver != nil {
				dbMsg, err := c.hub.messageSaver(context.Background(), &wsMsg)
				if err != nil {
					log.Printf("Failed to save message: %v", err)
					// 可以选择不广播，或者发送错误消息给客户端
					continue
				}
				// 更新消息的数据库 ID 和 SeqID
				if dbMsg != nil {
					wsMsg.MessageID = dbMsg.ID
					wsMsg.SeqID = dbMsg.SeqID
				}
			}
			// 保存成功后广播消息
			c.hub.broadcast <- &wsMsg
			continue
		}

		// 处理正在输入状态 (WS_TYPING)
		// 不存数据库，纯透传给聊天室其他成员
		if wsMsg.Type == WSTypingType {
			c.hub.broadcastToChatExcludeSender(&wsMsg, c.userID)
			continue
		}

		// 处理消息已读回执 (WS_MSG_READ)
		// 需要通过服务器确认并通知消息发送者
		if wsMsg.Type == WSMsgReadType {
			// TODO: 可以在这里调用 Service 更新已读状态
			// 或者让客户端调用 REST API，然后通过服务器通知
			// 这里直接广播给聊天室成员
			c.hub.broadcastToChatExcludeSender(&wsMsg, c.userID)
			continue
		}

		// 其他类型消息直接广播
		c.hub.broadcast <- &wsMsg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
