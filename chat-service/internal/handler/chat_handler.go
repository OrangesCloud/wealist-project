// internal/handler/chat_handler.go
package handler

import (
	"chat-service/internal/middleware"
	"chat-service/internal/model"
	"chat-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

type CreateChatRequest struct {
	WorkspaceID    string   `json:"workspaceId" binding:"required"`
	ProjectID      *string  `json:"projectId"`
	ChatType       string   `json:"chatType" binding:"required,oneof=DM GROUP PROJECT"`
	ChatName       string   `json:"chatName"`
	ParticipantIDs []string `json:"participantIds"`
}

type AddParticipantsRequest struct {
	UserIDs []string `json:"userIds" binding:"required,min=1"`
}

// CreateChat godoc
// @Summary      채팅방 생성
// @Description  새로운 채팅방을 생성합니다
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request body CreateChatRequest true "채팅방 생성 정보"
// @Success      201 {object} ChatResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /chats [post]
// @Security     BearerAuth
func (h *ChatHandler) CreateChat(c *gin.Context) {
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

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var projectID *uuid.UUID
	if req.ProjectID != nil {
		pid, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		projectID = &pid
	}

	// 참여자 ID 변환
	participantIDs := make([]uuid.UUID, 0, len(req.ParticipantIDs))
	for _, idStr := range req.ParticipantIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID: " + idStr})
			return
		}
		participantIDs = append(participantIDs, id)
	}

	chat, err := h.chatService.CreateChat(
		workspaceID,
		projectID,
		model.ChatType(req.ChatType),
		req.ChatName,
		userID,
		participantIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}


// GetChat godoc
// @Summary      채팅방 조회
// @Description  특정 채팅방의 상세 정보를 조회합니다
// @Tags         chat
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {object} ChatResponse
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /chats/{chatId} [get]
// @Security     BearerAuth
func (h *ChatHandler) GetChat(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	chat, err := h.chatService.GetChat(chatID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// GetWorkspaceChats godoc
// @Summary      워크스페이스 채팅방 목록
// @Description  워크스페이스의 모든 채팅방을 조회합니다
// @Tags         chat
// @Produce      json
// @Param        workspaceId path string true "Workspace ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {array} ChatResponse
// @Failure      400 {object} map[string]string
// @Router       /chats/workspace/{workspaceId} [get]
// @Security     BearerAuth
func (h *ChatHandler) GetWorkspaceChats(c *gin.Context) {
	workspaceID, err := uuid.Parse(c.Param("workspaceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	chats, err := h.chatService.GetWorkspaceChats(workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// GetMyChats godoc
// @Summary      내 채팅방 목록
// @Description  현재 사용자가 참여 중인 모든 채팅방을 조회합니다
// @Tags         chat
// @Produce      json
// @Success      200 {array} ChatResponse
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /chats/my [get]
// @Security     BearerAuth
func (h *ChatHandler) GetMyChats(c *gin.Context) {
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

	chats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// AddParticipants godoc
// @Summary      참여자 추가
// @Description  채팅방에 새로운 참여자를 추가합니다
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        request body AddParticipantsRequest true "추가할 사용자 ID 목록"
// @Success      200 {object} map[string]string "message: Participants added successfully"
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /chats/{chatId}/participants [post]
// @Security     BearerAuth
func (h *ChatHandler) AddParticipants(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	var req AddParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID: " + idStr})
			return
		}
		userIDs = append(userIDs, id)
	}

	if err := h.chatService.AddParticipants(chatID, userIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participants added successfully"})
}

// RemoveParticipant godoc
// @Summary      참여자 제거
// @Description  채팅방에서 참여자를 제거합니다
// @Tags         chat
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Param        userId path string true "User ID" example:"550e8400-e29b-41d4-a716-446655440001"
// @Success      200 {object} map[string]string "message: Participant removed successfully"
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /chats/{chatId}/participants/{userId} [delete]
// @Security     BearerAuth
func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.chatService.RemoveParticipant(chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant removed successfully"})
}

// DeleteChat godoc
// @Summary      채팅방 삭제
// @Description  채팅방을 삭제합니다 (Soft Delete)
// @Tags         chat
// @Produce      json
// @Param        chatId path string true "Chat ID" example:"550e8400-e29b-41d4-a716-446655440000"
// @Success      200 {object} map[string]string "message: Chat deleted successfully"
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /chats/{chatId} [delete]
// @Security     BearerAuth
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chatId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat ID"})
		return
	}

	if err := h.chatService.DeleteChat(chatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted successfully"})
}