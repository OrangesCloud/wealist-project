// internal/handler/message_handler.go
package handler

import (
	"chat-service/internal/middleware"
	"chat-service/internal/model"
	"chat-service/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	messageService service.MessageService
	chatService    service.ChatService
}

func NewMessageHandler(messageService service.MessageService, chatService service.ChatService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		chatService:    chatService,
	}
}

type SendMessageRequest struct {
	Content     string  `json:"content" binding:"required"`
	MessageType string  `json:"messageType" binding:"omitempty,oneof=TEXT IMAGE FILE"`
	FileURL     *string `json:"fileUrl"`
	FileName    *string `json:"fileName"`
	FileSize    *int64  `json:"fileSize"`
}

type MarkAsReadRequest struct {
	MessageIDs []string `json:"messageIds" binding:"required,min=1"`
}

// GetMessages godoc
// @Summary      메시지 히스토리 조회
// @Description  채팅방의 메시지 히스토리를 조회합니다 (페이지네이션)
// @Tags         message
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        limit query int false "페이지 크기 (기본: 50, 최대: 100)" default(50)
// @Param        offset query int false "오프셋 (기본: 0)" default(0)
// @Success      200 {array} handler.MessageResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/{chatId} [get]
// @Security     BearerAuth
func (h *MessageHandler) GetMessages(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	// 참여자인지 확인
	isParticipant, err := h.chatService.IsParticipant(chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant of this chat"})
		return
	}

	// 쿼리 파라미터 파싱
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.messageService.GetMessages(chatID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 DTO로 변환
	c.JSON(http.StatusOK, ToMessageResponses(messages))
}

// SendMessage godoc
// @Summary      메시지 전송
// @Description  채팅방에 메시지를 전송합니다 (REST fallback, WebSocket 권장)
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        request body SendMessageRequest true "메시지 내용"
// @Success      201 {object} handler.MessageResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/{chatId} [post]
// @Security     BearerAuth
func (h *MessageHandler) SendMessage(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	messageType := model.MessageTypeText
	if req.MessageType != "" {
		messageType = model.MessageType(req.MessageType)
	}

	message, err := h.messageService.CreateMessage(
		chatID,
		userID,
		req.Content,
		messageType,
		req.FileURL,
		req.FileName,
		req.FileSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 🔥 DTO로 변환
	c.JSON(http.StatusCreated, ToMessageResponse(message))
}

// DeleteMessage godoc
// @Summary      메시지 삭제
// @Description  메시지를 삭제합니다 (Soft Delete)
// @Tags         message
// @Produce      json
// @Param        messageId path string true "Message ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/{messageId} [delete]
// @Security     BearerAuth
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageID, err := uuid.Parse(c.Param("messageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.messageService.DeleteMessage(messageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}

// MarkMessagesAsRead godoc
// @Summary      메시지 읽음 처리
// @Description  여러 메시지를 읽음으로 처리합니다
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        request body MarkAsReadRequest true "읽은 메시지 ID 목록"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/read [post]
// @Security     BearerAuth
func (h *MessageHandler) MarkMessagesAsRead(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, msgIDStr := range req.MessageIDs {
		messageID, err := uuid.Parse(msgIDStr)
		if err != nil {
			continue
		}

		if err := h.messageService.MarkAsRead(messageID, userID); err != nil {
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Messages marked as read"})
}

// GetUnreadCount godoc
// @Summary      읽지 않은 메시지 수 조회
// @Description  채팅방의 읽지 않은 메시지 수를 조회합니다
// @Tags         message
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {object} map[string]int
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/{chatId}/unread [get]
// @Security     BearerAuth
func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	count, err := h.messageService.GetUnreadCount(chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// UpdateLastRead godoc
// @Summary      마지막 읽은 시간 업데이트
// @Description  채팅방의 마지막 읽은 시간을 현재 시간으로 업데이트합니다
// @Tags         message
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /messages/{chatId}/last-read [put]
// @Security     BearerAuth
func (h *MessageHandler) UpdateLastRead(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	if err := h.chatService.UpdateLastRead(chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Last read time updated"})
}