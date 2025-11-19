package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"project-board-api/internal/client"
	"project-board-api/internal/database"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func getLogger(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get("logger"); exists {
		if log, ok := logger.(*zap.Logger); ok {
			return log
		}
	}
	return zap.NewNop()
}

type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	BoardID string      `json:"boardId,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type WSHandler struct {
	Logger     *zap.Logger
	AuthClient client.UserClient
}

func NewWSHandler(log *zap.Logger, authClient client.UserClient) *WSHandler {
	return &WSHandler{
		Logger:     log,
		AuthClient: authClient,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[string]map[*Client]bool)
	clientsMu sync.RWMutex
)

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	projectID := c.Param("projectId")
	log := getLogger(c)

	log.Info("WebSocket connection attempt", zap.String("projectId", projectID))

	tokenStr := c.Query("token")
	if tokenStr == "" {
		log.Warn("WS connection attempt without token", zap.String("projectId", projectID))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	authCtx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	_, err := h.AuthClient.ValidateToken(authCtx, tokenStr)
	if err != nil {
		log.Error("WebSocket Auth Failed", zap.Error(err), zap.String("projectId", projectID))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Info("WebSocket auth successful", zap.String("projectId", projectID))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("WebSocket Upgrade Failed", zap.Error(err), zap.String("projectId", projectID))
		return
	}

	log.Info("WebSocket upgrade successful", zap.String("projectId", projectID))

	c.Abort()
	defer conn.Close()

	client := &Client{conn: conn, send: make(chan []byte, 256)}

	// 💡 [핵심 수정] 클라이언트 등록 전후 로그 추가
	clientsMu.Lock()
	if clients[projectID] == nil {
		clients[projectID] = make(map[*Client]bool)
		log.Info("Created new client map for project", zap.String("projectId", projectID))
	}
	clients[projectID][client] = true
	currentClientCount := len(clients[projectID])
	clientsMu.Unlock()

	log.Info("WebSocket client registered",
		zap.String("projectId", projectID),
		zap.Int("totalClients", currentClientCount))

	// 💡 [핵심 수정] writePump와 subscribeToRedis 고루틴 시작
	go h.writePump(client, projectID, log)
	go subscribeToRedis(projectID, client, log)

	// 연결 유지 및 끊김 처리
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Info("WebSocket connection closed",
				zap.String("projectId", projectID),
				zap.Error(err))

			clientsMu.Lock()
			delete(clients[projectID], client)
			if len(clients[projectID]) == 0 {
				delete(clients, projectID)
			}
			clientsMu.Unlock()
			close(client.send)
			break
		}
	}

	log.Info("WebSocket handler exiting", zap.String("projectId", projectID))
}

// 💡 [수정] writePump에 로그 추가
func (h *WSHandler) writePump(client *Client, projectID string, log *zap.Logger) {
	defer client.conn.Close()

	log.Info("writePump started", zap.String("projectId", projectID))

	for message := range client.send {
		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Error("WriteMessage failed", zap.Error(err), zap.String("projectId", projectID))
			return
		}
		log.Info("Message sent to client",
			zap.String("projectId", projectID),
			zap.String("message", string(message)))
	}

	log.Info("writePump exiting", zap.String("projectId", projectID))
}

// 💡 [수정] subscribeToRedis에 로그 추가
func subscribeToRedis(projectID string, client *Client, log *zap.Logger) {
	redis := database.GetRedis()
	pubsub := redis.Subscribe(context.Background(), "kanban:project:"+projectID)
	defer pubsub.Close()

	log.Info("Redis subscription started", zap.String("projectId", projectID))

	for msg := range pubsub.Channel() {
		log.Info("Redis message received",
			zap.String("projectId", projectID),
			zap.String("payload", msg.Payload))
		client.send <- []byte(msg.Payload)
	}

	log.Info("Redis subscription ended", zap.String("projectId", projectID))
}
func BroadcastEvent(projectID string, event WSEvent) {
	payload, _ := json.Marshal(event)

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	// 💡 [추가] 디버깅 로그
	fmt.Printf("🔊 [BROADCAST] ProjectID: %s, ClientCount: %d\n", projectID, len(clients[projectID]))
	fmt.Printf("🔊 [BROADCAST] Event: %+v\n", event)

	if projectClients, ok := clients[projectID]; ok {
		for client := range projectClients {
			select {
			case client.send <- payload:
				fmt.Println("✅ [BROADCAST] Message sent to client")
			default:
				fmt.Println("❌ [BROADCAST] Client channel full, closing")
				close(client.send)
			}
		}
	} else {
		fmt.Printf("⚠️ [BROADCAST] No clients found for project: %s\n", projectID)
	}

	redis := database.GetRedis()
	redis.Publish(context.Background(), "kanban:project:"+projectID, payload)
}
