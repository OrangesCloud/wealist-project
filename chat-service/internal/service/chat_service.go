// internal/service/chat_service.go
package service

import (
	"chat-service/internal/model"
	"chat-service/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type ChatService interface {
	CreateChat(workspaceID uuid.UUID, projectID *uuid.UUID, chatType model.ChatType, chatName string, createdBy uuid.UUID, participantIDs []uuid.UUID) (*model.Chat, error)
	GetChat(chatID uuid.UUID) (*model.Chat, error)
	GetWorkspaceChats(workspaceID uuid.UUID) ([]model.Chat, error)
	GetUserChats(userID uuid.UUID) ([]model.Chat, error)
	UpdateChat(chatID uuid.UUID, chatName string) error
	DeleteChat(chatID uuid.UUID) error
	
	AddParticipants(chatID uuid.UUID, userIDs []uuid.UUID) error
	RemoveParticipant(chatID, userID uuid.UUID) error
	GetParticipants(chatID uuid.UUID) ([]model.ChatParticipant, error)
	IsParticipant(chatID, userID uuid.UUID) (bool, error)
	UpdateLastRead(chatID, userID uuid.UUID) error
}

type chatService struct {
	chatRepo repository.ChatRepository
}

func NewChatService(chatRepo repository.ChatRepository) ChatService {
	return &chatService{
		chatRepo: chatRepo,
	}
}

func (s *chatService) CreateChat(
	workspaceID uuid.UUID,
	projectID *uuid.UUID,
	chatType model.ChatType,
	chatName string,
	createdBy uuid.UUID,
	participantIDs []uuid.UUID,
) (*model.Chat, error) {
	// 채팅방 생성
	chat := &model.Chat{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		ChatType:    chatType,
		ChatName:    chatName,
		CreatedBy:   createdBy,
	}

	if err := s.chatRepo.CreateChat(chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// 생성자를 참여자로 추가
	participantIDs = append([]uuid.UUID{createdBy}, participantIDs...)

	// 참여자 추가
	for _, userID := range participantIDs {
		participant := &model.ChatParticipant{
			ChatID: chat.ChatID,
			UserID: userID,
		}
		if err := s.chatRepo.AddParticipant(participant); err != nil {
			return nil, fmt.Errorf("failed to add participant: %w", err)
		}
	}

	return chat, nil
}

func (s *chatService) GetChat(chatID uuid.UUID) (*model.Chat, error) {
	return s.chatRepo.GetChatByID(chatID)
}

func (s *chatService) GetWorkspaceChats(workspaceID uuid.UUID) ([]model.Chat, error) {
	return s.chatRepo.GetChatsByWorkspace(workspaceID)
}

func (s *chatService) GetUserChats(userID uuid.UUID) ([]model.Chat, error) {
	return s.chatRepo.GetChatsByUser(userID)
}

func (s *chatService) UpdateChat(chatID uuid.UUID, chatName string) error {
	chat, err := s.chatRepo.GetChatByID(chatID)
	if err != nil {
		return err
	}

	chat.ChatName = chatName
	return s.chatRepo.UpdateChat(chat)
}

func (s *chatService) DeleteChat(chatID uuid.UUID) error {
	return s.chatRepo.DeleteChat(chatID)
}

func (s *chatService) AddParticipants(chatID uuid.UUID, userIDs []uuid.UUID) error {
	for _, userID := range userIDs {
		participant := &model.ChatParticipant{
			ChatID: chatID,
			UserID: userID,
		}
		if err := s.chatRepo.AddParticipant(participant); err != nil {
			return err
		}
	}
	return nil
}

func (s *chatService) RemoveParticipant(chatID, userID uuid.UUID) error {
	return s.chatRepo.RemoveParticipant(chatID, userID)
}

func (s *chatService) GetParticipants(chatID uuid.UUID) ([]model.ChatParticipant, error) {
	return s.chatRepo.GetParticipants(chatID)
}

func (s *chatService) IsParticipant(chatID, userID uuid.UUID) (bool, error) {
	return s.chatRepo.IsParticipant(chatID, userID)
}

func (s *chatService) UpdateLastRead(chatID, userID uuid.UUID) error {
	return s.chatRepo.UpdateLastRead(chatID, userID)
}