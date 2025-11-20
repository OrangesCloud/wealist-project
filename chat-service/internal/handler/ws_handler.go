// internal/handler/ws_handler.go
package handler

import (
	"chat-service/internal/client"
	"chat-service/internal/database"
	"chat-service/internal/model"
	"chat-service/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSMessage struct {
	Type        string                 `json:"type"`
	ChatID      string                 `json:"chatId,omitempty"`
	Content     string                 `json:"content,omitempty"`
	MessageType string                 `json:"messageType,omitempty"`
	FileURL     *string                `json:"fileUrl,omitempty"`
	FileName    *string                `json:"fileName,omitempty"`
	FileSize    *int64                 `json:"fileSize,omitempty"`
	MessageID   string                 `json:"messageId,omitempty"`
	UserID      string                 `json:"userId,omitempty"`
	Timestamp   time.Time              `json:"timestamp,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	chatID    uuid.UUID
	userID    uuid.UUID
	hub       *Hub
}

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	clientsMu  sync.RWMutex
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	logger     *zap.Logger
}

type WSHandler struct {
	logger         *zap.Logger
	userClient     client.UserClient
	messageService service.MessageService
	chatService    service.ChatService
	hub            *Hub
}

func NewWSHandler(
	logger *zap.Logger,
	userClient client.UserClient,
	messageService service.MessageService,
	chatService service.ChatService,
) *WSHandler {
	hub := &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		logger:     logger,
	}

	go hub.run()

	return &WSHandler{
		logger:         logger,
		userClient:     userClient,
		messageService: messageService,
		chatService:    chatService,
		hub:            hub,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			if h.clients[client.chatID] == nil {
				h.clients[client.chatID] = make(map[*Client]bool)
			}
			h.clients[client.chatID][client] = true
			h.clientsMu.Unlock()
			h.logger.Info("Client registered",
				zap.String("chatId", client.chatID.String()),
				zap.String("userId", client.userID.String()))

		case client := <-h.unregister:
			h.clientsMu.Lock()
			if clients, ok := h.clients[client.chatID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.chatID)
					}
				}
			}
			h.clientsMu.Unlock()
			h.logger.Info("Client unregistered",
				zap.String("chatId", client.chatID.String()),
				zap.String("userId", client.userID.String()))
		}
	}
}

func (h *Hub) broadcastToChat(chatID uuid.UUID, message []byte) {
	h.clientsMu.RLock()
	clients := h.clients[chatID]
	h.clientsMu.RUnlock()

	for client := range clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			h.unregister <- client
		}
	}
}

// HandleWebSocket godoc
// @Summary      WebSocket 연결
// @Description  채팅방 WebSocket에 연결합니다
// @Tags         websocket
// @Param        chatId path string true "Chat ID"
// @Param        token query string true "JWT Access Token"
// @Success      101 {string} string "Switching Protocols"
// @Failure      401 {object} map[string]string
// @Router       /ws/chat/{chatId} [get]
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	chatIDStr := c.Param("chatId")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	// 토큰 검증
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	validationResp, err := h.userClient.ValidateToken(ctx, token)
	if err != nil || !validationResp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID, err := uuid.Parse(validationResp.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 참여자인지 확인
	isParticipant, err := h.chatService.IsParticipant(chatID, userID)
	if err != nil || !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a participant"})
		return
	}

	// WebSocket 업그레이드
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := &Client{
		conn:   conn,
		send:   make(chan []byte, 256),
		chatID: chatID,
		userID: userID,
		hub:    h.hub,
	}

	h.hub.register <- client

	// Redis 구독 시작
	go h.subscribeToRedis(client)

	// Goroutines 시작
	go h.writePump(client)
	go h.readPump(client)
}

func (h *WSHandler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// 메시지 파싱
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			h.logger.Warn("Failed to parse message", zap.Error(err))
			continue
		}

		// 메시지 타입별 처리
		if err := h.handleMessage(client, &wsMsg); err != nil {
			h.logger.Error("Failed to handle message", zap.Error(err))
		}
	}
}

func (h *WSHandler) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) handleMessage(client *Client, wsMsg *WSMessage) error {
	switch wsMsg.Type {
	case "MESSAGE":
		return h.handleNewMessage(client, wsMsg)
	case "TYPING_START":
		return h.handleTyping(client, true)
	case "TYPING_STOP":
		return h.handleTyping(client, false)
	case "READ_MESSAGE":
		return h.handleReadMessage(client, wsMsg)
	default:
		h.logger.Warn("Unknown message type", zap.String("type", wsMsg.Type))
	}
	return nil
}

func (h *WSHandler) handleNewMessage(client *Client, wsMsg *WSMessage) error {
	messageType := model.MessageTypeText
	if wsMsg.MessageType != "" {
		messageType = model.MessageType(wsMsg.MessageType)
	}

	message, err := h.messageService.CreateMessage(
		client.chatID,
		client.userID,
		wsMsg.Content,
		messageType,
		wsMsg.FileURL,
		wsMsg.FileName,
		wsMsg.FileSize,
	)
	if err != nil {
		return err
	}

	// 브로드캐스트는 CreateMessage 내부에서 처리됨
	h.logger.Info("Message created via WebSocket",
		zap.String("messageId", message.MessageID.String()),
		zap.String("chatId", client.chatID.String()))

	return nil
}

func (h *WSHandler) handleTyping(client *Client, isTyping bool) error {
	eventType := "USER_TYPING_STOP"
	if isTyping {
		eventType = "USER_TYPING"
	}

	payload, _ := json.Marshal(WSMessage{
		Type:   eventType,
		ChatID: client.chatID.String(),
		UserID: client.userID.String(),
	})

	h.hub.broadcastToChat(client.chatID, payload)
	return nil
}

func (h *WSHandler) handleReadMessage(client *Client, wsMsg *WSMessage) error {
	if wsMsg.MessageID == "" {
		return fmt.Errorf("messageId required")
	}

	messageID, err := uuid.Parse(wsMsg.MessageID)
	if err != nil {
		return err
	}

	if err := h.messageService.MarkAsRead(messageID, client.userID); err != nil {
		return err
	}

	// 읽음 알림 브로드캐스트
	payload, _ := json.Marshal(WSMessage{
		Type:      "MESSAGE_READ",
		MessageID: wsMsg.MessageID,
		UserID:    client.userID.String(),
		ChatID:    client.chatID.String(),
		Timestamp: time.Now(),
	})

	h.hub.broadcastToChat(client.chatID, payload)
	return nil
}

func (h *WSHandler) subscribeToRedis(client *Client) {
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Recovered from panic in subscribeToRedis",
				zap.Any("panic", r),
				zap.String("chatId", client.chatID.String()))
		}
	}()

	pubsub := database.SubscribeChatEvents(client.chatID.String())
	if pubsub == nil {
		h.logger.Warn("Redis pubsub not available")
		return
	}
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		select {
		case client.send <- []byte(msg.Payload):
		case <-time.After(1 * time.Second):
			h.logger.Warn("Failed to send Redis message to client",
				zap.String("chatId", client.chatID.String()))
			return
		}
	}
}