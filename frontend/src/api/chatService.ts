// src/api/chat/chatService.ts

import { chatServiceClient } from './apiConfig';
import { AxiosResponse } from 'axios';
import type { Chat, Message, CreateChatRequest, SendMessageRequest } from '../types/chat';

/**
 * 🔥 DM 채팅방 생성 또는 기존 채팅방 가져오기
 * @param targetUserId 대화 상대방 userId
 * @param workspaceId 워크스페이스 ID
 * @returns chatId
 */
export const createOrGetDMChat = async (
  targetUserId: string,
  workspaceId: string,
): Promise<string> => {
  try {
    console.log('🔍 DM 채팅방 찾기/생성:', { targetUserId, workspaceId });

    // 1. 내 채팅방 목록 가져오기
    const myChats = await getMyChats();
    console.log('📋 내 채팅방 목록:', myChats.length, '개');

    // 2. 이미 존재하는 DM 채팅방 찾기
    const existingDM = myChats.find((chat) => {
      if (chat.chatType !== 'DM') return false;

      // participants가 있으면 확인
      if (chat.participants) {
        const participantUserIds = chat.participants.map((p) => p.userId);
        return participantUserIds.includes(targetUserId);
      }

      return false;
    });

    if (existingDM) {
      console.log('✅ 기존 DM 채팅방 사용:', existingDM.chatId);
      return existingDM.chatId;
    }

    // 3. 없으면 새로 생성
    console.log('🆕 새 DM 채팅방 생성 중...');
    const newChat = await createChat({
      workspaceId,
      chatType: 'DM',
      participantIds: [targetUserId],
    });

    console.log('✅ 새 채팅방 생성 완료:', newChat.chatId);
    return newChat.chatId;
  } catch (error) {
    console.error('❌ Failed to create or get DM chat:', error);
    throw error;
  }
};

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

// ============================================================================
// 🔥 Presence API (온라인 상태)
// ============================================================================

/**
 * 온라인 사용자 목록 조회
 * [API] GET /api/chats/presence/online
 */
export const getOnlineUsers = async (): Promise<string[]> => {
  const response: AxiosResponse<{ onlineUsers: string[]; count: number }> =
    await chatServiceClient.get('/presence/online');
  return response.data.onlineUsers;
};

/**
 * 특정 사용자 온라인 여부 확인
 * [API] GET /api/chats/presence/status/{userId}
 * @param userId 확인할 사용자 ID
 * @returns true: 온라인, false: 오프라인
 */
export const checkUserStatus = async (userId: string): Promise<boolean> => {
  const response: AxiosResponse<{ userId: string; isOnline: boolean }> =
    await chatServiceClient.get(`/presence/status/${userId}`);
  return response.data.isOnline;
};

/**
 * 여러 사용자의 온라인 상태 일괄 확인
 * @param userIds 확인할 사용자 ID 배열
 * @returns Map<userId, isOnline>
 */
export const checkMultipleUserStatus = async (userIds: string[]): Promise<Map<string, boolean>> => {
  try {
    // 온라인 사용자 목록 가져오기
    const onlineUsers = await getOnlineUsers();
    const onlineSet = new Set(onlineUsers);

    // Map으로 변환
    const statusMap = new Map<string, boolean>();
    userIds.forEach((userId) => {
      statusMap.set(userId, onlineSet.has(userId));
    });

    return statusMap;
  } catch (error) {
    console.error('Failed to check multiple user status:', error);
    // 에러 시 모두 오프라인으로 처리
    const statusMap = new Map<string, boolean>();
    userIds.forEach((userId) => {
      statusMap.set(userId, false);
    });
    return statusMap;
  }
};
