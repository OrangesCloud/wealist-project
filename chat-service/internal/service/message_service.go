package service

import (
	"chat-service/internal/database"
	"chat-service/internal/model"
	"chat-service/internal/repository"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageService interface {
	CreateMessage(chatID, userID uuid.UUID, content string, messageType model.MessageType, fileURL, fileName *string, fileSize *int64) (*model.Message, error)
	GetMessage(messageID uuid.UUID) (*model.Message, error)
	GetMessages(chatID uuid.UUID, limit, offset int) ([]model.Message, error)
	GetMessagesAfter(chatID uuid.UUID, after time.Time, limit int) ([]model.Message, error)
	UpdateMessage(messageID uuid.UUID, content string) error
	DeleteMessage(messageID uuid.UUID) error
	
	MarkAsRead(messageID, userID uuid.UUID) error
	GetUnreadCount(chatID, userID uuid.UUID) (int64, error)
	GetMessageReads(messageID uuid.UUID) ([]model.MessageRead, error)
	
	BroadcastMessage(message *model.Message) error
}

type messageService struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
}

func NewMessageService(messageRepo repository.MessageRepository, chatRepo repository.ChatRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
	}
}

func (s *messageService) CreateMessage(
	chatID, userID uuid.UUID,
	content string,
	messageType model.MessageType,
	fileURL, fileName *string,
	fileSize *int64,
) (*model.Message, error) {
	// 참여자인지 확인
	isParticipant, err := s.chatRepo.IsParticipant(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, fmt.Errorf("user is not a participant of this chat")
	}

	// 메시지 생성
	message := &model.Message{
		ChatID:      chatID,
		UserID:      userID,
		Content:     content,
		MessageType: messageType,
		FileURL:     fileURL,
		FileName:    fileName,
		FileSize:    fileSize,
	}

	if err := s.messageRepo.CreateMessage(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Redis Pub/Sub으로 브로드캐스트
	if err := s.BroadcastMessage(message); err != nil {
		// 로그만 남기고 에러는 무시 (메시지는 저장됨)
		fmt.Printf("Warning: failed to broadcast message: %v\n", err)
	}

	return message, nil
}

func (s *messageService) GetMessage(messageID uuid.UUID) (*model.Message, error) {
	return s.messageRepo.GetMessageByID(messageID)
}

func (s *messageService) GetMessages(chatID uuid.UUID, limit, offset int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.GetMessagesByChatID(chatID, limit, offset)
}

func (s *messageService) GetMessagesAfter(chatID uuid.UUID, after time.Time, limit int) ([]model.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	return s.messageRepo.GetMessagesAfter(chatID, after, limit)
}

func (s *messageService) UpdateMessage(messageID uuid.UUID, content string) error {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	message.Content = content
	return s.messageRepo.UpdateMessage(message)
}

func (s *messageService) DeleteMessage(messageID uuid.UUID) error {
	return s.messageRepo.DeleteMessage(messageID)
}

func (s *messageService) MarkAsRead(messageID, userID uuid.UUID) error {
	return s.messageRepo.MarkAsRead(messageID, userID)
}

func (s *messageService) GetUnreadCount(chatID, userID uuid.UUID) (int64, error) {
	return s.messageRepo.GetUnreadCount(chatID, userID)
}

func (s *messageService) GetMessageReads(messageID uuid.UUID) ([]model.MessageRead, error) {
	return s.messageRepo.GetMessageReads(messageID)
}

func (s *messageService) BroadcastMessage(message *model.Message) error {
	payload, err := json.Marshal(map[string]interface{}{
		"type":      "MESSAGE_RECEIVED",
		"messageId": message.MessageID,
		"chatId":    message.ChatID,
		"userId":    message.UserID,
		"content":   message.Content,
		"messageType": message.MessageType,
		"fileUrl":   message.FileURL,
		"fileName":  message.FileName,
		"fileSize":  message.FileSize,
		"createdAt": message.CreatedAt,
	})
	if err != nil {
		return err
	}

	return database.PublishChatEvent(message.ChatID.String(), payload)
}