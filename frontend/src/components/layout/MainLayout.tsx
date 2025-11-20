// src/components/layout/MainLayout.tsx

import React, { useState, useEffect } from 'react';
import { useTheme } from '../../contexts/ThemeContext';
import { UserProfileResponse } from '../../types/user';
import { getMyProfile } from '../../api/userService';
import { Sidebar } from './Sidebar';
import { UserMenu } from './UserMenu';
import { ChatManager } from '../chat/ChatManager';

interface MainLayoutProps {
  onLogout: () => void;
  workspaceId: string;
  projectId?: string;
  children: React.ReactNode;
  onProfileModalOpen: () => void;
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
        onChatToggle={() => setIsChatOpen(!isChatOpen)}
        onUserMenuToggle={() => setShowUserMenu(!showUserMenu)}
      />

      {/* 메인 콘텐츠 */}
      <main
        className="flex-grow flex flex-col relative z-10"
        style={{
          marginLeft: sidebarWidth,
          marginRight: isChatOpen ? '20rem' : '0',
          transition: 'margin-right 0.3s ease',
          minHeight: '100vh',
        }}
      >
        {children}
      </main>

      {/* 채팅 관리자 */}
      <ChatManager
        workspaceId={workspaceId}
        projectId={projectId}
        isOpen={isChatOpen}
        onClose={() => setIsChatOpen(false)}
      />

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
    </div>
  );
};

export default MainLayout;
