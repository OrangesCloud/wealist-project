// src/hooks/useChatWebSocket.ts

import { useEffect, useRef, useCallback } from 'react';
import {
  connectChatWebSocket,
  disconnectChatWebSocket,
  sendChatMessage,
  isChatWebSocketConnected,
  WS_CHAT_MTH,
} from '../utils/chatWebsocket';

interface UseChatWebSocketOptions {
  chatId: string | null;
  onMessage: (data: any) => void;
  autoConnect?: boolean;
}

/**
 * Chat WebSocket Hook
 *
 * @example
 * const { sendMessage, isConnected } = useChatWebSocket({
 *   chatId: 'chat-123',
 *   onMessage: (event) => {
 *     if (event.type === 'MESSAGE_RECEIVED') {
 *       setMessages(prev => [...prev, event.payload]);
 *     }
 *   }
 * });
 */
export const useChatWebSocket = ({
  chatId,
  onMessage,
  autoConnect = true,
}: UseChatWebSocketOptions) => {
  const wsConnectedRef = useRef(false);
  const onMessageRef = useRef(onMessage);

  // onMessage 최신 상태 유지
  useEffect(() => {
    onMessageRef.current = onMessage;
  }, [onMessage]);

  // WebSocket 연결
  useEffect(() => {
    if (!chatId || !autoConnect) return;

    // 🔥 1번만 연결
    if (!wsConnectedRef.current) {
      wsConnectedRef.current = true;
      console.log('🔌 [Chat WS Hook] 연결 시작:', chatId);

      connectChatWebSocket(chatId, (event) => {
        console.log('🔊 [Chat WS Hook] 이벤트 수신:', event);

        // 채팅 이벤트만 필터링
        if (WS_CHAT_MTH.includes(event.type)) {
          onMessageRef.current(event);
        }
      });
    }

    // 🔥 클린업: 언마운트 시에만 연결 해제
    return () => {
      if (wsConnectedRef.current) {
        console.log('🔌 [Chat WS Hook] 컴포넌트 언마운트 - 연결 해제');
        disconnectChatWebSocket();
        wsConnectedRef.current = false;
      }
    };
  }, [chatId, autoConnect]);

  // 메시지 전송 (텍스트)
  const sendMessage = useCallback((content: string) => {
    return sendChatMessage({
      type: 'MESSAGE',
      content,
      messageType: 'TEXT',
    });
  }, []);

  // 메시지 전송 (파일 포함)
  const sendFileMessage = useCallback((content: string, fileData: any) => {
    return sendChatMessage({
      type: 'MESSAGE',
      content,
      messageType: fileData.messageType || 'FILE',
      fileUrl: fileData.fileUrl,
      fileName: fileData.fileName,
      fileSize: fileData.fileSize,
    });
  }, []);

  // 타이핑 상태 전송
  const sendTyping = useCallback((isTyping: boolean) => {
    return sendChatMessage({
      type: isTyping ? 'TYPING_START' : 'TYPING_STOP',
    });
  }, []);

  // 읽음 처리
  const markAsRead = useCallback((messageId: string) => {
    return sendChatMessage({
      type: 'READ_MESSAGE',
      messageId,
    });
  }, []);

  return {
    sendMessage,
    sendFileMessage,
    sendTyping,
    markAsRead,
    isConnected: isChatWebSocketConnected(),
  };
};
