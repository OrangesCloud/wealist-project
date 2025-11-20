// src/components/chat/ChatPanel.tsx

import React, { useState, useEffect, useRef } from 'react';
import { useChatWebSocket } from '../../hooks/useChatWebsocket';
import { getMessages } from '../../api/chatService';
import type { Message } from '../../types/chat';

interface ChatPanelProps {
  chatId: string;
  onClose: () => void;
}

export const ChatPanel: React.FC<ChatPanelProps> = ({ chatId, onClose }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // 🔥 WebSocket 연결
  const { sendMessage, sendTyping, isConnected } = useChatWebSocket({
    chatId,
    onMessage: (event) => {
      console.log('🔊 [ChatPanel] 이벤트 수신:', event);

      if (event.type === 'MESSAGE_RECEIVED') {
        setMessages((prev) => [...prev, event.payload]);
      }

      if (event.type === 'USER_TYPING') {
        // 타이핑 인디케이터 표시
        console.log('⌨️ User typing:', event.userId);
      }
    },
  });

  // 메시지 로드
  useEffect(() => {
    const loadMessages = async () => {
      setIsLoading(true);
      try {
        const msgs = await getMessages(chatId);
        setMessages(msgs);
      } catch (error) {
        console.error('Failed to load messages:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadMessages();
  }, [chatId]);

  // 자동 스크롤
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // 메시지 전송
  const handleSendMessage = () => {
    if (!inputMessage.trim()) return;

    const success = sendMessage(inputMessage);
    if (success) {
      setInputMessage('');
    }
  };

  // 타이핑 인디케이터
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputMessage(e.target.value);
    sendTyping(true);

    // 1초 후 타이핑 중지
    setTimeout(() => sendTyping(false), 1000);
  };

  return (
    <div className="fixed top-0 right-0 h-full w-80 bg-white shadow-2xl flex flex-col z-40">
      {/* 헤더 */}
      <div className="p-4 border-b">
        <div className="flex items-center justify-between">
          <h3 className="font-bold">채팅</h3>
          <button onClick={onClose}>✕</button>
        </div>
        <div className="text-xs text-gray-500 mt-1">
          {isConnected ? '🟢 연결됨' : '🔴 연결 끊김'}
        </div>
      </div>

      {/* 메시지 영역 */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {isLoading ? (
          <div className="flex justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
          </div>
        ) : (
          messages.map((msg) => (
            <div
              key={msg.messageId}
              className={`flex ${msg.isMine ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-[70%] rounded-lg p-3 ${
                  msg.isMine ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-900'
                }`}
              >
                {!msg.isMine && <p className="text-xs font-bold mb-1 opacity-70">{msg.userName}</p>}
                <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                <p className={`text-xs mt-1 ${msg.isMine ? 'text-blue-100' : 'text-gray-500'}`}>
                  {new Date(msg.createdAt).toLocaleTimeString('ko-KR', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </p>
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* 입력 영역 */}
      <div className="p-4 border-t">
        <div className="flex items-center gap-2">
          <input
            type="text"
            value={inputMessage}
            onChange={handleInputChange}
            onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
            placeholder="메시지를 입력하세요..."
            className="flex-1 p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button
            onClick={handleSendMessage}
            disabled={!inputMessage.trim()}
            className="p-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:bg-gray-300"
          >
            전송
          </button>
        </div>
      </div>
    </div>
  );
};
