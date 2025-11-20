// src/components/layout/MainLayout.tsx

import React, { useState, useEffect } from 'react';
import { useTheme } from '../../contexts/ThemeContext';
import { UserProfileResponse, WorkspaceMemberResponse } from '../../types/user';
import { getMyProfile } from '../../api/userService';
import { createOrGetDMChat } from '../../api/chatService';
import { Sidebar } from './Sidebar';
import { UserMenu } from './UserMenu';
import { ChatPanel } from '../chat/chatPanel';
import { ChatListPanel } from '../chat/ChatListPanel';

interface MainLayoutProps {
  onLogout: () => void;
  workspaceId: string;
  projectId?: string;
  children: React.ReactNode;
  onProfileModalOpen: () => void;
  onStartChat?: (member: WorkspaceMemberResponse) => void;
}

const MainLayout: React.FC<MainLayoutProps> = ({
  onLogout,
  workspaceId,
  projectId,
  children,
  onProfileModalOpen,
}) => {
  const { theme } = useTheme();

  // States
  const [userProfile, setUserProfile] = useState<UserProfileResponse | null>(null);
  const [isLoadingProfile, setIsLoadingProfile] = useState(true);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [isChatOpen, setIsChatOpen] = useState(false);
  const [activeChatId, setActiveChatId] = useState<string | null>(null);
  const [isLoadingChat, setIsLoadingChat] = useState(false);

  const sidebarWidth = 'w-16 sm:w-20';

  // 프로필 로드
  useEffect(() => {
    const fetchUserProfile = async () => {
      try {
        const profile = await getMyProfile();
        setUserProfile(profile);
      } catch (e) {
        console.error('기본 프로필 로드 실패:', e);
      } finally {
        setIsLoadingProfile(false);
      }
    };
    fetchUserProfile();
  }, []);

  // 🔥 채팅 시작 핸들러
  const handleStartChat = async (member: WorkspaceMemberResponse) => {
    setIsLoadingChat(true);
    try {
      console.log('🔵 채팅 시작:', member.userName);

      // 1. DM 채팅방 생성 또는 기존 채팅방 가져오기
      const chatId = await createOrGetDMChat(member.userId, workspaceId);
      console.log('✅ 채팅방 ID:', chatId);

      // 2. ChatPanel 열기
      setActiveChatId(chatId);
      setIsChatOpen(true);
    } catch (error) {
      console.error('❌ Failed to start chat:', error);
      alert('채팅방을 열 수 없습니다.');
    } finally {
      setIsLoadingChat(false);
    }
  };

  // 외부 클릭 감지 (UserMenu)
  useEffect(() => {
    if (!showUserMenu) return;

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (!target.closest('[data-user-menu]')) {
        setShowUserMenu(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [showUserMenu]);

  if (isLoadingProfile) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
      </div>
    );
  }

  return (
    <div className={`min-h-screen flex ${theme.colors.background} relative`}>
      {/* 백그라운드 패턴 */}
      <div
        className="fixed inset-0 opacity-5"
        style={{
          backgroundImage:
            'linear-gradient(#000 1px, transparent 1px), linear-gradient(90deg, #000 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }}
      />

      {/* 사이드바 */}
      <Sidebar
        workspaceId={workspaceId}
        userProfile={userProfile}
        isChatActive={isChatOpen}
        onChatToggle={() => {
          setIsChatOpen(!isChatOpen);
          if (isChatOpen) {
            setActiveChatId(null); // 채팅 닫을 때 활성 채팅 초기화
          }
        }}
        onUserMenuToggle={() => setShowUserMenu(!showUserMenu)}
        onStartChat={handleStartChat}
      />

      {/* 메인 콘텐츠 */}
      <main
        className="flex-grow flex flex-col relative z-10"
        style={{
          marginLeft: sidebarWidth,
          marginRight: isChatOpen && activeChatId ? '20rem' : '0',
          transition: 'margin-right 0.3s ease',
          minHeight: '100vh',
        }}
      >
        {children}
      </main>

      {/* 🔥 ChatPanel 또는 ChatList */}
      {isChatOpen && (
        <>
          {activeChatId ? (
            // 특정 채팅방 열림
            <ChatPanel
              chatId={activeChatId}
              onClose={() => {
                setActiveChatId(null);
                setIsChatOpen(false);
              }}
              onBack={() => setActiveChatId(null)} // 뒤로가기 → 채팅 리스트로
            />
          ) : (
            // 채팅 리스트
            <ChatListPanel
              workspaceId={workspaceId}
              onChatSelect={(chatId) => setActiveChatId(chatId)}
              onClose={() => setIsChatOpen(false)}
            />
          )}
        </>
      )}

      {/* 유저 메뉴 */}
      {showUserMenu && (
        <div data-user-menu>
          <UserMenu
            userProfile={userProfile}
            onProfileModalOpen={onProfileModalOpen}
            onLogout={onLogout}
            onClose={() => setShowUserMenu(false)}
          />
        </div>
      )}

      {/* 🔥 채팅 로딩 오버레이 */}
      {isLoadingChat && (
        <div className="fixed inset-0 bg-black/20 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 shadow-xl">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto" />
            <p className="mt-3 text-sm text-gray-600">채팅방을 여는 중...</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default MainLayout;
