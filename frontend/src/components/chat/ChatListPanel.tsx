// src/components/chat/ChatListPanel.tsx

import React, { useState, useEffect } from 'react';
import { X, Search, MessageCircle, Users, User } from 'lucide-react';
import { getMyChats } from '../../api/chatService';
import type { Chat } from '../../types/chat';

interface ChatListPanelProps {
  workspaceId: string;
  onChatSelect: (chatId: string) => void;
  onClose: () => void;
}

export const ChatListPanel: React.FC<ChatListPanelProps> = ({
  workspaceId,
  onChatSelect,
  onClose,
}) => {
  const [chats, setChats] = useState<Chat[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // 채팅방 목록 로드
  useEffect(() => {
    const loadChats = async () => {
      setIsLoading(true);
      try {
        const allChats = await getMyChats();
        // 워크스페이스 필터링 (옵션)
        const filteredChats = allChats.filter((chat) => chat.workspaceId === workspaceId);
        setChats(filteredChats);
      } catch (error) {
        console.error('Failed to load chats:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadChats();
  }, [workspaceId]);

  // 검색 필터링
  const filteredChats = chats.filter((chat) =>
    chat.chatName?.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  // 채팅 타입별 아이콘
  const getChatIcon = (chatType: string) => {
    switch (chatType) {
      case 'DM':
        return <User className="w-5 h-5 text-blue-500" />;
      case 'GROUP':
        return <Users className="w-5 h-5 text-green-500" />;
      case 'PROJECT':
        return <MessageCircle className="w-5 h-5 text-purple-500" />;
      default:
        return <MessageCircle className="w-5 h-5 text-gray-500" />;
    }
  };

  // 마지막 메시지 시간 포맷
  const formatTime = (date: string) => {
    const now = new Date();
    const messageDate = new Date(date);
    const diffMs = now.getTime() - messageDate.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return '방금 전';
    if (diffMins < 60) return `${diffMins}분 전`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}시간 전`;
    return messageDate.toLocaleDateString('ko-KR', { month: 'short', day: 'numeric' });
  };

  return (
    <div className="h-full w-full bg-white flex flex-col">
      {/* 헤더 */}
      <div className="p-4 border-b bg-gradient-to-r from-blue-600 to-blue-700 text-white">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <MessageCircle className="w-5 h-5" />
            <h2 className="font-bold text-lg">채팅</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-white/20 rounded transition"
            title="닫기"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 검색바 */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="채팅방 검색..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-white/20 text-white placeholder-white/60 rounded-lg focus:outline-none focus:ring-2 focus:ring-white/50"
          />
        </div>
      </div>

      {/* 채팅 리스트 */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
          </div>
        ) : filteredChats.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-gray-400">
            <MessageCircle className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">{searchQuery ? '검색 결과가 없습니다' : '채팅방이 없습니다'}</p>
            <p className="text-xs mt-1">멤버 아바타를 클릭해서 채팅을 시작하세요</p>
          </div>
        ) : (
          <div className="divide-y">
            {filteredChats.map((chat) => (
              <button
                key={chat.chatId}
                onClick={() => onChatSelect(chat.chatId)}
                className="w-full p-4 hover:bg-gray-50 transition text-left"
              >
                <div className="flex items-start gap-3">
                  {/* 아이콘 */}
                  <div className="flex-shrink-0 mt-1">{getChatIcon(chat.chatType)}</div>

                  {/* 내용 */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between mb-1">
                      <h3 className="font-semibold text-sm text-gray-900 truncate">
                        {chat.chatName || `${chat.chatType} 채팅`}
                      </h3>
                      <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
                        {formatTime(chat.updatedAt)}
                      </span>
                    </div>

                    <div className="flex items-center justify-between">
                      <p className="text-xs text-gray-500 truncate">
                        {chat.chatType === 'DM' && '1:1 대화'}
                        {chat.chatType === 'GROUP' && `그룹 채팅`}
                        {chat.chatType === 'PROJECT' && `프로젝트 채팅`}
                      </p>

                      {/* 읽지 않은 메시지 뱃지 */}
                      {chat.unreadCount && chat.unreadCount > 0 && (
                        <span className="flex-shrink-0 ml-2 min-w-[20px] h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center px-1.5 font-bold">
                          {chat.unreadCount > 99 ? '99+' : chat.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
