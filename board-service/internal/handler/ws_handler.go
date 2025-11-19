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

// ============================================================================
// 💡 [추가] WebSocket 타임아웃 설정
// ============================================================================
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10  // 54초마다 ping
	maxMessageSize = 512
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
	conn      *websocket.Conn
	send      chan []byte
	projectID string  // 💡 [추가] 프로젝트 ID 저장
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

// HandleWebSocket godoc
// @Summary      WebSocket 실시간 연결
// @Description  프로젝트의 실시간 이벤트를 구독하기 위한 WebSocket 연결을 설정합니다
// @Description  연결 후 BOARD_CREATED, BOARD_UPDATED, BOARD_MOVED, BOARD_DELETED 이벤트를 실시간으로 수신합니다
// @Description  인증은 쿼리 파라미터로 전달된 JWT 토큰을 통해 수행됩니다
// @Tags         websocket
// @Produce      json
// @Param        projectId path string true "Project ID (UUID)"
// @Param        token query string true "JWT Access Token"
// @Success      101 {string} string "Switching Protocols - WebSocket 연결 성공"
// @Failure      401 {object} project-board-api_internal_response.ErrorResponse "인증 실패"
// @Failure      500 {object} project-board-api_internal_response.ErrorResponse "서버 에러"
// @Router       /ws/project/{projectId} [get]
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

	// 🔥 [중요] c.Abort()를 Upgrade 전에 호출하면 안 됨!
	// c.Abort() ← 이거 삭제!
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("WebSocket Upgrade Failed", zap.Error(err), zap.String("projectId", projectID))
		return
	}

	log.Info("WebSocket upgrade successful", zap.String("projectId", projectID))
	defer conn.Close()

	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 256),
		projectID: projectID,
	}

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

	go h.writePump(client, log)
	go h.readPump(client, log)
	go subscribeToRedis(projectID, client, log)

	// 🔥 [추가] 메인 고루틴 대기 (readPump가 끝날 때까지)
	select {}
}
// ============================================================================
// 💡 [신규] readPump: 클라이언트로부터 메시지 수신 + Pong 처리
// ============================================================================
func (h *WSHandler) readPump(client *Client, log *zap.Logger) {
	defer func() {
		log.Info("🔌 readPump: Client disconnected", zap.String("projectId", client.projectID))
		
		// 클라이언트 등록 해제
		clientsMu.Lock()
		delete(clients[client.projectID], client)
		if len(clients[client.projectID]) == 0 {
			delete(clients, client.projectID)
		}
		clientsMu.Unlock()
		
		close(client.send)
		client.conn.Close()
	}()

	// 타임아웃 설정
	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	
	// 🔥 Pong 핸들러 등록
	client.conn.SetPongHandler(func(string) error {
		log.Info("🏓 Pong received from client", zap.String("projectId", client.projectID))
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	log.Info("readPump started", zap.String("projectId", client.projectID))

	for {
		messageType, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("❌ Unexpected WebSocket close", zap.Error(err), zap.String("projectId", client.projectID))
			} else {
				log.Info("⚠️ Normal WebSocket close", zap.String("projectId", client.projectID))
			}
			break
		}

		// 💡 [추가] Ping 메시지 처리 (클라이언트가 JSON으로 보낼 경우)
		if messageType == websocket.TextMessage {
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err == nil {
				if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
					log.Info("🏓 Ping received (JSON), sending pong", zap.String("projectId", client.projectID))
					// Pong 응답 전송
					pongMsg, _ := json.Marshal(map[string]string{"type": "pong"})
					select {
					case client.send <- pongMsg:
					default:
						log.Warn("⚠️ Client send channel full", zap.String("projectId", client.projectID))
					}
					continue
				}
			}
		}

		log.Info("📨 Message received from client",
			zap.String("projectId", client.projectID),
			zap.Int("type", int(messageType)),
			zap.Int("length", len(message)))
	}

	log.Info("readPump exiting", zap.String("projectId", client.projectID))
}

// ============================================================================
// 💡 [수정] writePump: Ping 전송 + 메시지 전송
// ============================================================================
func (h *WSHandler) writePump(client *Client, log *zap.Logger) {
	// 🔥 Ping 타이머 생성
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
		log.Info("writePump exiting", zap.String("projectId", client.projectID))
	}()

	log.Info("writePump started", zap.String("projectId", client.projectID))

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 채널이 닫힘
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Error("❌ WriteMessage failed", zap.Error(err), zap.String("projectId", client.projectID))
				return
			}
			log.Info("✅ Message sent to client",
				zap.String("projectId", client.projectID),
				zap.String("message", string(message)))

		// 🔥 Ping 메시지 주기적 전송
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("❌ Ping failed", zap.Error(err), zap.String("projectId", client.projectID))
				return
			}
			log.Info("🏓 Ping sent to client", zap.String("projectId", client.projectID))
		}
	}
}

// ============================================================================
// subscribeToRedis: Redis Pub/Sub 구독
// ============================================================================
func subscribeToRedis(projectID string, client *Client, log *zap.Logger) {
	redis := database.GetRedis()
	pubsub := redis.Subscribe(context.Background(), "kanban:project:"+projectID)
	defer pubsub.Close()

	log.Info("Redis subscription started", zap.String("projectId", projectID))

	for msg := range pubsub.Channel() {
		log.Info("Redis message received",
			zap.String("projectId", projectID),
			zap.String("payload", msg.Payload))
		
		select {
		case client.send <- []byte(msg.Payload):
			log.Info("✅ Redis message sent to client", zap.String("projectId", projectID))
		default:
			log.Warn("⚠️ Client send channel full, message dropped", zap.String("projectId", projectID))
		}
	}

	log.Info("Redis subscription ended", zap.String("projectId", projectID))
}

// ============================================================================
// BroadcastEvent: 이벤트 브로드캐스트
// ============================================================================
func BroadcastEvent(projectID string, event WSEvent) {
	payload, _ := json.Marshal(event)

	clientsMu.RLock()
	defer clientsMu.RUnlock()

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