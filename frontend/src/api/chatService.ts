// src/api/chat/chatService.ts

import { chatServiceClient } from './apiConfig';
import { AxiosResponse } from 'axios';
import type { Chat, Message, CreateChatRequest, SendMessageRequest } from '../types/chat';

/**
 * 채팅방 생성
 * [API] POST /api/chats
 */
export const createChat = async (data: CreateChatRequest): Promise<Chat> => {
  const response: AxiosResponse<Chat> = await chatServiceClient.post('/chats', data);
  return response.data;
};

/**
 * 내 채팅방 목록 조회
 * [API] GET /api/chats/my
 */
export const getMyChats = async (): Promise<Chat[]> => {
  const response: AxiosResponse<Chat[]> = await chatServiceClient.get('/chats/my');
  return response.data;
};

/**
 * 워크스페이스 채팅방 목록 조회
 * [API] GET /api/chats/workspace/{workspaceId}
 */
export const getWorkspaceChats = async (workspaceId: string): Promise<Chat[]> => {
  const response: AxiosResponse<Chat[]> = await chatServiceClient.get(
    `/chats/workspace/${workspaceId}`,
  );
  return response.data;
};

/**
 * 프로젝트 채팅방 조회 (필터링)
 * 워크스페이스의 모든 채팅방 중 특정 프로젝트 채팅만 필터링
 */
export const getProjectChats = async (projectId: string): Promise<Chat[]> => {
  const chats = await getMyChats();
  return chats.filter((chat) => chat.projectId === projectId || chat.chatType === 'PROJECT');
};

/**
 * 채팅방 상세 조회
 * [API] GET /api/chats/{chatId}
 */
export const getChat = async (chatId: string): Promise<Chat> => {
  const response: AxiosResponse<Chat> = await chatServiceClient.get(`/chats/${chatId}`);
  return response.data;
};

/**
 * 채팅방 삭제
 * [API] DELETE /api/chats/{chatId}
 */
export const deleteChat = async (chatId: string): Promise<void> => {
  await chatServiceClient.delete(`/chats/${chatId}`);
};

/**
 * 참여자 추가
 * [API] POST /api/chats/{chatId}/participants
 */
export const addParticipants = async (chatId: string, userIds: string[]): Promise<void> => {
  await chatServiceClient.post(`/chats/${chatId}/participants`, { userIds });
};

/**
 * 참여자 제거
 * [API] DELETE /api/chats/{chatId}/participants/{userId}
 */
export const removeParticipant = async (chatId: string, userId: string): Promise<void> => {
  await chatServiceClient.delete(`/chats/${chatId}/participants/${userId}`);
};

/**
 * 메시지 히스토리 조회
 * [API] GET /api/chats/messages/{chatId}
 */
export const getMessages = async (chatId: string, limit = 50, offset = 0): Promise<Message[]> => {
  const response: AxiosResponse<Message[]> = await chatServiceClient.get(`/messages/${chatId}`, {
    params: { limit, offset },
  });

  // 🔥 현재 사용자 ID 가져오기
  const currentUserId = localStorage.getItem('userId');

  // isMine 플래그 추가
  return response.data.map((msg) => ({
    ...msg,
    isMine: msg.userId === currentUserId,
  }));
};

/**
 * 메시지 전송 (REST fallback)
 * [API] POST /api/chats/messages/{chatId}
 */
export const sendMessage = async (chatId: string, content: string): Promise<Message> => {
  const data: SendMessageRequest = { content };
  const response: AxiosResponse<Message> = await chatServiceClient.post(
    `/messages/${chatId}`,
    data,
  );
  return response.data;
};

/**
 * 메시지 삭제
 * [API] DELETE /api/chats/messages/{messageId}
 */
export const deleteMessage = async (messageId: string): Promise<void> => {
  await chatServiceClient.delete(`/messages/${messageId}`);
};

/**
 * 메시지 읽음 처리
 * [API] POST /api/chats/messages/read
 */
export const markMessagesAsRead = async (messageIds: string[]): Promise<void> => {
  await chatServiceClient.post('/messages/read', { messageIds });
};

/**
 * 읽지 않은 메시지 수 조회
 * [API] GET /api/chats/messages/{chatId}/unread
 */
export const getUnreadCount = async (chatId: string): Promise<number> => {
  const response: AxiosResponse<{ unreadCount: number }> = await chatServiceClient.get(
    `/messages/${chatId}/unread`,
  );
  return response.data.unreadCount;
};

/**
 * 마지막 읽은 시간 업데이트
 * [API] PUT /api/chats/messages/{chatId}/last-read
 */
export const updateLastRead = async (chatId: string): Promise<void> => {
  await chatServiceClient.put(`/messages/${chatId}/last-read`);
};
